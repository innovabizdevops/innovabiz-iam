-- Seeds hierárquicos para todas as tabelas de localização

INSERT INTO geographies.country (id, nome, codigo_iso) VALUES (uuid_generate_v4(), 'Angola', 'AO');
INSERT INTO geographies.province (id, nome, country_id) SELECT uuid_generate_v4(), 'Luanda', id FROM geographies.country WHERE codigo_iso = 'AO';
INSERT INTO geographies.state (id, nome, country_id, province_id) SELECT uuid_generate_v4(), 'Estado Central', c.id, p.id FROM geographies.country c JOIN geographies.province p ON p.country_id = c.id WHERE c.codigo_iso = 'AO';
INSERT INTO geographies.municipality (id, nome, state_id, province_id) SELECT uuid_generate_v4(), 'Município A', s.id, p.id FROM geographies.state s JOIN geographies.province p ON s.province_id = p.id;
INSERT INTO geographies.district (id, nome, municipality_id, state_id) SELECT uuid_generate_v4(), 'Distrito 1', m.id, s.id FROM geographies.municipality m JOIN geographies.state s ON m.state_id = s.id;
INSERT INTO geographies.county (id, nome, district_id, municipality_id) SELECT uuid_generate_v4(), 'Concelho 1', d.id, m.id FROM geographies.district d JOIN geographies.municipality m ON d.municipality_id = m.id;
INSERT INTO geographies.parish (id, nome, county_id) SELECT uuid_generate_v4(), 'Freguesia 1', c.id FROM geographies.county c;
INSERT INTO geographies.commune (id, nome, municipality_id) SELECT uuid_generate_v4(), 'Comuna 1', m.id FROM geographies.municipality m;
INSERT INTO geographies.neighborhood (id, nome, district_id, municipality_id, commune_id) SELECT uuid_generate_v4(), 'Bairro 1', d.id, m.id, co.id FROM geographies.district d JOIN geographies.municipality m ON d.municipality_id = m.id JOIN geographies.commune co ON co.municipality_id = m.id;
INSERT INTO geographies.city (id, nome, state_id, municipality_id, district_id) SELECT uuid_generate_v4(), 'Cidade 1', s.id, m.id, d.id FROM geographies.state s JOIN geographies.municipality m ON m.state_id = s.id JOIN geographies.district d ON d.state_id = s.id;
