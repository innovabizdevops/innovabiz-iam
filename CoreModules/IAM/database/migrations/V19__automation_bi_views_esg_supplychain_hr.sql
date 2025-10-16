-- V19__automation_bi_views_esg_supplychain_hr.sql
-- Triggers, funções e views para automação, auditoria e BI dos domínios ESG, Supply Chain e RH

-- Função genérica para atualização de updated_at
CREATE OR REPLACE FUNCTION update_timestamp()
RETURNS TRIGGER AS $$
BEGIN
  NEW.updated_at = now();
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- ESG: Trigger de atualização de timestamp
DROP TRIGGER IF EXISTS trg_update_esg_company_report_timestamp ON esg.company_esg_report;
CREATE TRIGGER trg_update_esg_company_report_timestamp
BEFORE UPDATE ON esg.company_esg_report
FOR EACH ROW EXECUTE FUNCTION update_timestamp();

-- Supply Chain: Trigger de atualização de timestamp
DROP TRIGGER IF EXISTS trg_update_supplychain_contract_timestamp ON supply_chain.logistics_contract;
CREATE TRIGGER trg_update_supplychain_contract_timestamp
BEFORE UPDATE ON supply_chain.logistics_contract
FOR EACH ROW EXECUTE FUNCTION update_timestamp();

-- RH: Trigger de atualização de timestamp
DROP TRIGGER IF EXISTS trg_update_hr_employee_timestamp ON hr.employee;
CREATE TRIGGER trg_update_hr_employee_timestamp
BEFORE UPDATE ON hr.employee
FOR EACH ROW EXECUTE FUNCTION update_timestamp();

-- Função de auditoria genérica
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

-- Triggers de auditoria
DROP TRIGGER IF EXISTS trg_audit_esg_company_report ON esg.company_esg_report;
CREATE TRIGGER trg_audit_esg_company_report
AFTER UPDATE OR DELETE ON esg.company_esg_report
FOR EACH ROW EXECUTE FUNCTION audit_log();

DROP TRIGGER IF EXISTS trg_audit_supplychain_contract ON supply_chain.logistics_contract;
CREATE TRIGGER trg_audit_supplychain_contract
AFTER UPDATE OR DELETE ON supply_chain.logistics_contract
FOR EACH ROW EXECUTE FUNCTION audit_log();

DROP TRIGGER IF EXISTS trg_audit_hr_employee ON hr.employee;
CREATE TRIGGER trg_audit_hr_employee
AFTER UPDATE OR DELETE ON hr.employee
FOR EACH ROW EXECUTE FUNCTION audit_log();

-- Views BI ESG
CREATE OR REPLACE VIEW bi.vw_esg_companies_missing_reports AS
SELECT c.id AS company_id, c.name AS company_name, y.year
FROM organization.company c
CROSS JOIN (SELECT DISTINCT year FROM esg.company_esg_report) y
LEFT JOIN esg.company_esg_report r ON r.company_id = c.id AND r.year = y.year
WHERE r.id IS NULL;

-- Views BI Supply Chain
CREATE OR REPLACE VIEW bi.vw_suppliers_missing_audit AS
SELECT s.*
FROM supply_chain.supplier s
LEFT JOIN (
    SELECT supplier_id, MAX(audit_date) AS last_audit FROM supply_chain.supplier_audit GROUP BY supplier_id
) la ON la.supplier_id = s.id
WHERE la.last_audit IS NULL OR la.last_audit < (now() - INTERVAL '1 year');

-- Views BI RH
CREATE OR REPLACE VIEW bi.vw_employees_missing_indicators AS
SELECT e.id AS employee_id, e.first_name, e.last_name, y.year
FROM hr.employee e
CROSS JOIN (SELECT DISTINCT year FROM hr.employee_hr_report) y
LEFT JOIN hr.employee_hr_report r ON r.employee_id = e.id AND r.year = y.year
WHERE r.id IS NULL;

-- Replicar triggers/views conforme necessidade para outros indicadores e tabelas dos domínios expandidos.
