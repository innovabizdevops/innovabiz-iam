-- Exemplo de triggers de auditoria para todas as tabelas de localização

CREATE OR REPLACE FUNCTION geographies_audit_func() RETURNS trigger AS $$
BEGIN
  IF TG_OP = 'UPDATE' THEN
    INSERT INTO audit_log(table_name, operation, record_id, changed_at)
    VALUES (TG_TABLE_NAME, 'UPDATE', NEW.id, now());
    RETURN NEW;
  ELSIF TG_OP = 'DELETE' THEN
    INSERT INTO audit_log(table_name, operation, record_id, changed_at)
    VALUES (TG_TABLE_NAME, 'DELETE', OLD.id, now());
    RETURN OLD;
  END IF;
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Triggers para cada tabela
DO $$
DECLARE
  tbl TEXT;
BEGIN
  FOREACH tbl IN ARRAY ARRAY[
    'country','province','state','municipality','district','county','parish','commune','neighborhood','city'
  ]
  LOOP
    EXECUTE format('CREATE TRIGGER %I_audit_trg AFTER UPDATE OR DELETE ON geographies.%I FOR EACH ROW EXECUTE FUNCTION geographies_audit_func();', tbl, tbl);
  END LOOP;
END$$;
