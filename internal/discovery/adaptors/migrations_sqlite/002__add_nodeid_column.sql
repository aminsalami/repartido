--migrate:up
ALTER TABLE cache_node ADD COLUMN node_id text not null default '';
CREATE UNIQUE INDEX node_id_unique ON cache_node(node_id);
