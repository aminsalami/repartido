--migrate:up
create table if not exists cache_node
(
    id        integer PRIMARY KEY,
    name      varchar(100) not null,
    host      varchar(100) not null,
    port      integer      not null,

    last_ping datetime     not null,
    ram_size  int          not null
);