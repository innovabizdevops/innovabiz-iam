-- Seeds para o domínio de geografias

INSERT INTO geographies.country (id, nome, codigo_iso)
VALUES
  (uuid_generate_v4(), 'Brasil', 'BR'),
  (uuid_generate_v4(), 'Portugal', 'PT'),
  (uuid_generate_v4(), 'Estados Unidos', 'US');

INSERT INTO geographies.state (id, nome, codigo_uf, country_id)
SELECT uuid_generate_v4(), 'São Paulo', 'SP', c.id FROM geographies.country c WHERE c.codigo_iso = 'BR';
INSERT INTO geographies.state (id, nome, codigo_uf, country_id)
SELECT uuid_generate_v4(), 'Lisboa', NULL, c.id FROM geographies.country c WHERE c.codigo_iso = 'PT';
INSERT INTO geographies.state (id, nome, codigo_uf, country_id)
SELECT uuid_generate_v4(), 'California', 'CA', c.id FROM geographies.country c WHERE c.codigo_iso = 'US';

INSERT INTO geographies.city (id, nome, state_id)
SELECT uuid_generate_v4(), 'São Paulo', s.id FROM geographies.state s WHERE s.nome = 'São Paulo';
INSERT INTO geographies.city (id, nome, state_id)
SELECT uuid_generate_v4(), 'Lisboa', s.id FROM geographies.state s WHERE s.nome = 'Lisboa';
INSERT INTO geographies.city (id, nome, state_id)
SELECT uuid_generate_v4(), 'San Francisco', s.id FROM geographies.state s WHERE s.nome = 'California';
