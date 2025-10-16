-- V15__governance_automation.sql
-- Triggers, funções e views para automação de auditoria, compliance e integração BI

-- Função para atualização automática de timestamps
CREATE OR REPLACE FUNCTION update_timestamp()
RETURNS TRIGGER AS $$
BEGIN
  NEW.updated_at = now();
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Exemplo: Trigger para update de timestamps em organization.company
DROP TRIGGER IF EXISTS trg_update_company_timestamp ON organization.company;
CREATE TRIGGER trg_update_company_timestamp
BEFORE UPDATE ON organization.company
FOR EACH ROW EXECUTE FUNCTION update_timestamp();

-- Função de auditoria para log de alterações
CREATE OR REPLACE FUNCTION audit_log()
RETURNS TRIGGER AS $$
BEGIN
  INSERT INTO public.audit_log (
    table_name, record_id, operation, changed_by, changed_at, old_data, new_data
  ) VALUES (
    TG_TABLE_NAME, NEW.id, TG_OP, NEW.updated_by, now(), row_to_json(OLD), row_to_json(NEW)
  );
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Tabela de log de auditoria
CREATE TABLE IF NOT EXISTS public.audit_log (
  id SERIAL PRIMARY KEY,
  table_name VARCHAR(100),
  record_id UUID,
  operation VARCHAR(10),
  changed_by UUID,
  changed_at TIMESTAMP,
  old_data JSONB,
  new_data JSONB
);

-- Exemplo: Trigger de auditoria em organization.company
DROP TRIGGER IF EXISTS trg_audit_company ON organization.company;
CREATE TRIGGER trg_audit_company
AFTER UPDATE OR DELETE ON organization.company
FOR EACH ROW EXECUTE FUNCTION audit_log();

-- Trigger de compliance: alerta se registro for inserido/atualizado sem compliance_status ou responsible
CREATE OR REPLACE FUNCTION notify_missing_compliance()
RETURNS TRIGGER AS $$
BEGIN
  IF NEW.compliance_status IS NULL OR NEW.compliance_status = ''
     OR NEW.responsible IS NULL OR NEW.responsible = '' THEN
    RAISE NOTICE 'Registro sem compliance_status ou responsible: %', NEW.id;
  END IF;
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Exemplo: Trigger de compliance em organization.company
DROP TRIGGER IF EXISTS trg_notify_company_compliance ON organization.company;
CREATE TRIGGER trg_notify_company_compliance
BEFORE INSERT OR UPDATE ON organization.company
FOR EACH ROW EXECUTE FUNCTION notify_missing_compliance();

-- Views integradas para BI e auditoria
CREATE OR REPLACE VIEW bi.vw_companies_missing_compliance AS
SELECT c.*
FROM organization.company c
WHERE (c.compliance_status IS NULL OR c.compliance_status = ''
   OR c.responsible IS NULL OR c.responsible = '');

CREATE OR REPLACE VIEW bi.vw_risk_critical AS
SELECT r.*, rl.name AS risk_level_name
FROM risk.risk r
JOIN risk.risk_level rl ON r.risk_level_id = rl.id
WHERE rl.name ILIKE '%crítico%' OR rl.name ILIKE '%critical%';

-- Replicar triggers e views para outros domínios conforme padrão acima.
-- Adicione triggers de timestamp, auditoria e compliance para as principais tabelas dos domínios organization, business, products, services, risk, compliance, processes.
-- Adicione views de integração BI para gaps de compliance, riscos críticos, processos sem responsável, etc.

-- Comentários multilíngues e documentação automática devem ser gerados via script auxiliar/documentação.
