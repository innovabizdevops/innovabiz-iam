-- Exemplo de trigger de auditoria para tabela compliance.norma
CREATE OR REPLACE FUNCTION compliance_audit_norma() RETURNS trigger AS $$
BEGIN
    IF TG_OP = 'UPDATE' THEN
        INSERT INTO compliance.audit_log (
            entidade, entidade_id, acao, usuario_id, data_acao, valor_anterior, valor_novo, origem
        ) VALUES (
            'norma', NEW.id, 'UPDATE', current_user::uuid, now(), row_to_json(OLD), row_to_json(NEW), 'trigger'
        );
    ELSIF TG_OP = 'DELETE' THEN
        INSERT INTO compliance.audit_log (
            entidade, entidade_id, acao, usuario_id, data_acao, valor_anterior, valor_novo, origem
        ) VALUES (
            'norma', OLD.id, 'DELETE', current_user::uuid, now(), row_to_json(OLD), NULL, 'trigger'
        );
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER tr_audit_norma
AFTER UPDATE OR DELETE ON compliance.norma
FOR EACH ROW EXECUTE FUNCTION compliance_audit_norma();
