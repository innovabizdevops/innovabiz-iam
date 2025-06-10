-- View para governança integrada de especialistas humanos e digitais
CREATE OR REPLACE VIEW vw_especialistas_governanca AS
SELECT
    e.nome AS especialista,
    e.funcao,
    e.area_atuacao,
    e.certificacoes,
    e.contato,
    e.disponibilidade,
    v.code AS validador_digital,
    v.name AS nome_validador,
    v.validator_class,
    v.framework_id,
    v.is_active AS validador_ativo,
    v.version AS versao_validador
FROM equipe_especialistas e
LEFT JOIN compliance_validators v ON v.owner = e.nome;
