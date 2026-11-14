-- +goose Up
CREATE TABLE netlist (
    list TEXT NOT NULL CHECK (list IN ('white', 'black')),
    cidr CIDR NOT NULL,
    PRIMARY KEY (list, cidr)
);

-- +goose Down
DROP TABLE netlist;
