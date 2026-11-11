package netlist

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/netip"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"

	"github.com/s0meuserhere/antibrutforce/internal/domain/netlist"
	"github.com/s0meuserhere/antibrutforce/internal/domain/subnet"
)

// cacheKey - ключ кэша обоих списков в Redis.
const cacheKey = "netlist"

var errUnknownList = errors.New("unknown list")

// Repository хранит whitelist и blacklist в PostgreSQL и кэширует их в Redis.
type Repository struct {
	db       *pgxpool.Pool
	cache    *redis.Client
	cacheTTL time.Duration
}

// rNetlist - строка таблицы netlist. В кэше Redis лежат такие же строки.
type rNetlist struct {
	List netlist.Kind `db:"list" json:"list"`
	CIDR netip.Prefix `db:"cidr" json:"cidr"`
}

// New возвращает репозиторий списков. cacheTTL ограничивает, насколько кэш
// может отстать от PostgreSQL, если запись в базу и чтение промаха пересеклись.
func New(db *pgxpool.Pool, cache *redis.Client, cacheTTL time.Duration) *Repository {
	return &Repository{db: db, cache: cache, cacheTTL: cacheTTL}
}

// Get возвращает whitelist и blacklist. При попадании в кэш база не читается,
// при промахе списки читаются из PostgreSQL и кладутся в кэш.
func (r *Repository) Get(ctx context.Context) (netlist.Lists, error) {
	rows, hit, err := r.readCache(ctx)
	if err != nil {
		return netlist.Lists{}, err
	}

	if !hit {
		if rows, err = r.readDB(ctx); err != nil {
			return netlist.Lists{}, err
		}

		if err := r.writeCache(ctx, rows); err != nil {
			return netlist.Lists{}, err
		}
	}

	var whitelist, blacklist []subnet.Subnet

	for _, row := range rows {
		switch row.List {
		case netlist.Whitelist:
			whitelist = append(whitelist, subnet.Restore(row.CIDR))
		case netlist.Blacklist:
			blacklist = append(blacklist, subnet.Restore(row.CIDR))
		default:
			return netlist.Lists{}, fmt.Errorf("%w: %q", errUnknownList, row.List)
		}
	}

	return netlist.New(whitelist, blacklist), nil
}

// Add добавляет подсеть в список. Повторное добавление ничего не меняет.
func (r *Repository) Add(ctx context.Context, e netlist.Entry) error {
	query := "INSERT INTO netlist (list, cidr) VALUES ($1, $2) ON CONFLICT DO NOTHING"
	err := r.write(ctx, e, query)

	return err
}

// Remove удаляет подсеть из списка. Удаление отсутствующей подсети ничего не меняет.
func (r *Repository) Remove(ctx context.Context, e netlist.Entry) error {
	query := "DELETE FROM netlist WHERE list = $1 AND cidr = $2"
	err := r.write(ctx, e, query)

	return err
}

// write выполняет запись в PostgreSQL и удаляет кэш списков.
func (r *Repository) write(ctx context.Context, e netlist.Entry, query string) error {
	if _, err := r.db.Exec(ctx, query, string(e.Kind()), e.Subnet().Prefix()); err != nil {
		return fmt.Errorf("write %s list: %w", e.Kind(), err)
	}

	if err := r.cache.Del(ctx, cacheKey).Err(); err != nil {
		return fmt.Errorf("drop lists cache: %w", err)
	}

	return nil
}

func (r *Repository) readCache(ctx context.Context) ([]rNetlist, bool, error) {
	data, err := r.cache.Get(ctx, cacheKey).Bytes()
	if errors.Is(err, redis.Nil) {
		return nil, false, nil
	}

	if err != nil {
		return nil, false, fmt.Errorf("read lists cache: %w", err)
	}

	var rows []rNetlist
	if err := json.Unmarshal(data, &rows); err != nil {
		return nil, false, fmt.Errorf("decode lists cache: %w", err)
	}

	return rows, true, nil
}

func (r *Repository) readDB(ctx context.Context) ([]rNetlist, error) {
	query := "SELECT list, cidr FROM netlist ORDER BY list, cidr"

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("query lists: %w", err)
	}

	entries, err := pgx.CollectRows(rows, pgx.RowToStructByName[rNetlist])
	if err != nil {
		return nil, fmt.Errorf("scan lists: %w", err)
	}

	return entries, nil
}

func (r *Repository) writeCache(ctx context.Context, rows []rNetlist) error {
	data, err := json.Marshal(rows)
	if err != nil {
		return fmt.Errorf("encode lists cache: %w", err)
	}

	if err := r.cache.Set(ctx, cacheKey, data, r.cacheTTL).Err(); err != nil {
		return fmt.Errorf("fill lists cache: %w", err)
	}

	return nil
}
