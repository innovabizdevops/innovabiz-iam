"""
Modelos de dados para integração com Bureau de Créditos.

Autor: Eduardo Jeremias
Projeto: INNOVABIZ IAM/TrustGuard
Data: 21/08/2025
"""

from datetime import datetime
from enum import Enum
from typing import Any, Dict, List, Optional, Set, Union
from pydantic import BaseModel, Field, validator


class ScoreLevel(str, Enum):
    """Níveis de score de crédito"""
    VERY_LOW = "very_low"       # Muito baixo
    LOW = "low"                 # Baixo
    MEDIUM = "medium"           # Médio
    HIGH = "high"               # Alto
    VERY_HIGH = "very_high"     # Muito alto


class RiskLevel(str, Enum):
    """Níveis de risco de crédito"""
    VERY_LOW = "very_low"       # Muito baixo
    LOW = "low"                 # Baixo
    MEDIUM = "medium"           # Médio
    HIGH = "high"               # Alto
    VERY_HIGH = "very_high"     # Muito alto


class DebtStatus(str, Enum):
    """Status de dívidas"""
    PAID = "paid"               # Paga
    OVERDUE = "overdue"         # Vencida
    IN_DAY = "in_day"           # Em dia
    RENEGOTIATED = "renegotiated"  # Renegociada
    DISPUTED = "disputed"       # Contestada


class FraudRiskLevel(str, Enum):
    """Níveis de risco de fraude"""
    VERY_LOW = "very_low"       # Muito baixo
    LOW = "low"                 # Baixo
    MEDIUM = "medium"           # Médio
    HIGH = "high"               # Alto
    VERY_HIGH = "very_high"     # Muito alto


class VerificationStatus(str, Enum):
    """Status de verificação de identidade"""
    VERIFIED = "verified"       # Verificado
    PARTIALLY_VERIFIED = "partially_verified"  # Parcialmente verificado
    NOT_VERIFIED = "not_verified"  # Não verificado
    INCONCLUSIVE = "inconclusive"  # Inconclusivo
    ERROR = "error"             # Erro


class TransactionType(str, Enum):
    """Tipos de transação"""
    CREDIT = "credit"           # Crédito
    DEBIT = "debit"             # Débito
    LOAN = "loan"               # Empréstimo
    PAYMENT = "payment"         # Pagamento
    TRANSFER = "transfer"       # Transferência
    OTHER = "other"             # Outro


class CreditScore(BaseModel):
    """Informações de score de crédito"""
    document: str = Field(..., description="Número do documento (CPF/CNPJ)")
    score: int = Field(..., description="Score de crédito (0-1000)")
    score_level: ScoreLevel = Field(..., description="Nível do score de crédito")
    provider: str = Field(..., description="Provedor do bureau de crédito")
    date: datetime = Field(..., description="Data da consulta")
    previous_score: Optional[int] = Field(None, description="Score anterior")
    previous_date: Optional[datetime] = Field(None, description="Data do score anterior")
    trend: Optional[str] = Field(None, description="Tendência (up, down, stable)")
    percentile: Optional[int] = Field(None, description="Percentil na população")
    factors: List[Dict[str, Any]] = Field(default_factory=list, 
                                          description="Fatores que influenciam o score")
    metadata: Dict[str, Any] = Field(default_factory=dict, description="Metadados adicionais")
    
    @validator("score")
    def validate_score(cls, v):
        if not 0 <= v <= 1000:
            raise ValueError("Score must be between 0 and 1000")
        return v


class Debt(BaseModel):
    """Informações sobre dívidas"""
    id: str = Field(..., description="Identificador da dívida")
    creditor: str = Field(..., description="Nome do credor")
    original_amount: float = Field(..., description="Valor original da dívida")
    current_amount: float = Field(..., description="Valor atual da dívida")
    due_date: datetime = Field(..., description="Data de vencimento")
    contract_date: datetime = Field(..., description="Data do contrato")
    status: DebtStatus = Field(..., description="Status da dívida")
    description: Optional[str] = Field(None, description="Descrição da dívida")
    payment_history: Optional[List[Dict[str, Any]]] = Field(None, 
                                                           description="Histórico de pagamentos")
    contract_type: Optional[str] = Field(None, description="Tipo de contrato")
    last_update: datetime = Field(..., description="Data da última atualização")


class Protest(BaseModel):
    """Informações sobre protestos"""
    id: str = Field(..., description="Identificador do protesto")
    value: float = Field(..., description="Valor do protesto")
    date: datetime = Field(..., description="Data do protesto")
    city: str = Field(..., description="Cidade do cartório")
    state: str = Field(..., description="Estado do cartório")
    notary_office: str = Field(..., description="Cartório")
    status: str = Field(..., description="Status do protesto")
    creditor: str = Field(..., description="Nome do credor")


class LegalProcess(BaseModel):
    """Informações sobre processos judiciais"""
    id: str = Field(..., description="Número do processo")
    court: str = Field(..., description="Tribunal")
    jurisdiction: str = Field(..., description="Jurisdição")
    nature: str = Field(..., description="Natureza do processo")
    value: Optional[float] = Field(None, description="Valor do processo")
    date: datetime = Field(..., description="Data de início")
    status: str = Field(..., description="Status do processo")
    plaintiff: Optional[str] = Field(None, description="Autor")
    defendant: Optional[str] = Field(None, description="Réu")
    last_update: datetime = Field(..., description="Data da última atualização")


class CreditReport(BaseModel):
    """Relatório de crédito completo"""
    document: str = Field(..., description="Número do documento (CPF/CNPJ)")
    name: str = Field(..., description="Nome da pessoa ou empresa")
    provider: str = Field(..., description="Provedor do bureau de crédito")
    date: datetime = Field(..., description="Data da consulta")
    score: CreditScore = Field(..., description="Score de crédito")
    debts: List[Debt] = Field(default_factory=list, description="Lista de dívidas")
    protests: List[Protest] = Field(default_factory=list, description="Lista de protestos")
    legal_processes: List[LegalProcess] = Field(default_factory=list, 
                                               description="Processos judiciais")
    total_debt_amount: float = Field(..., description="Valor total de dívidas")
    debt_count: int = Field(..., description="Quantidade de dívidas")
    overdue_debt_count: int = Field(..., description="Quantidade de dívidas vencidas")
    credit_queries_last_30days: int = Field(..., 
                                           description="Consultas de crédito nos últimos 30 dias")
    credit_queries_last_12months: int = Field(..., 
                                            description="Consultas de crédito nos últimos 12 meses")
    bank_accounts: List[Dict[str, Any]] = Field(default_factory=list, description="Contas bancárias")
    credit_cards: List[Dict[str, Any]] = Field(default_factory=list, description="Cartões de crédito")
    loans: List[Dict[str, Any]] = Field(default_factory=list, description="Empréstimos")
    report_id: str = Field(..., description="ID do relatório")
    metadata: Dict[str, Any] = Field(default_factory=dict, description="Metadados adicionais")


class CreditRisk(BaseModel):
    """Avaliação de risco de crédito"""
    document: str = Field(..., description="Número do documento (CPF/CNPJ)")
    name: Optional[str] = Field(None, description="Nome da pessoa ou empresa")
    risk_level: RiskLevel = Field(..., description="Nível de risco de crédito")
    risk_score: int = Field(..., description="Score de risco (0-1000, maior = mais risco)")
    provider: str = Field(..., description="Provedor do bureau de crédito")
    date: datetime = Field(..., description="Data da avaliação")
    risk_factors: List[Dict[str, Any]] = Field(default_factory=list, 
                                              description="Fatores de risco identificados")
    recommendation: str = Field(..., description="Recomendação (aprovar, negar, revisar)")
    transaction_amount: Optional[float] = Field(None, description="Valor da transação analisada")
    transaction_type: Optional[str] = Field(None, description="Tipo da transação analisada")
    likelihood_of_default: float = Field(..., description="Probabilidade de inadimplência")
    confidence: float = Field(..., description="Nível de confiança da avaliação")
    metadata: Dict[str, Any] = Field(default_factory=dict, description="Metadados adicionais")
    
    @validator("risk_score")
    def validate_risk_score(cls, v):
        if not 0 <= v <= 1000:
            raise ValueError("Risk score must be between 0 and 1000")
        return v
    
    @validator("likelihood_of_default", "confidence")
    def validate_probability(cls, v):
        if not 0 <= v <= 1:
            raise ValueError("Probability must be between 0 and 1")
        return v


class IdentityVerification(BaseModel):
    """Resultado da verificação de identidade"""
    document: str = Field(..., description="Número do documento (CPF/CNPJ)")
    name: str = Field(..., description="Nome da pessoa ou empresa")
    status: VerificationStatus = Field(..., description="Status da verificação")
    provider: str = Field(..., description="Provedor do bureau de crédito")
    date: datetime = Field(..., description="Data da verificação")
    score: int = Field(..., description="Score de verificação (0-1000)")
    verified_fields: Dict[str, bool] = Field(default_factory=dict, 
                                            description="Campos verificados e status")
    verification_methods: List[str] = Field(default_factory=list, 
                                          description="Métodos de verificação utilizados")
    verification_id: str = Field(..., description="ID da verificação")
    metadata: Dict[str, Any] = Field(default_factory=dict, description="Metadados adicionais")
    
    @validator("score")
    def validate_score(cls, v):
        if not 0 <= v <= 1000:
            raise ValueError("Verification score must be between 0 and 1000")
        return v


class FraudIndicator(BaseModel):
    """Indicadores de fraude"""
    document: str = Field(..., description="Número do documento (CPF/CNPJ)")
    name: Optional[str] = Field(None, description="Nome da pessoa ou empresa")
    risk_level: FraudRiskLevel = Field(..., description="Nível de risco de fraude")
    risk_score: int = Field(..., description="Score de risco (0-1000, maior = mais risco)")
    provider: str = Field(..., description="Provedor do bureau de crédito")
    date: datetime = Field(..., description="Data da avaliação")
    indicators: List[Dict[str, Any]] = Field(default_factory=list, 
                                           description="Indicadores de fraude identificados")
    ip_address: Optional[str] = Field(None, description="Endereço IP analisado")
    device_id: Optional[str] = Field(None, description="ID do dispositivo analisado")
    geolocation: Optional[Dict[str, Any]] = Field(None, description="Informações de geolocalização")
    velocity_checks: Optional[Dict[str, Any]] = Field(None, 
                                                    description="Verificações de velocidade")
    recommendation: str = Field(..., description="Recomendação (aprovar, negar, revisar)")
    confidence: float = Field(..., description="Nível de confiança da avaliação")
    metadata: Dict[str, Any] = Field(default_factory=dict, description="Metadados adicionais")
    
    @validator("risk_score")
    def validate_risk_score(cls, v):
        if not 0 <= v <= 1000:
            raise ValueError("Risk score must be between 0 and 1000")
        return v
    
    @validator("confidence")
    def validate_confidence(cls, v):
        if not 0 <= v <= 1:
            raise ValueError("Confidence must be between 0 and 1")
        return v


class Transaction(BaseModel):
    """Transação financeira"""
    id: str = Field(..., description="Identificador da transação")
    date: datetime = Field(..., description="Data da transação")
    type: TransactionType = Field(..., description="Tipo da transação")
    amount: float = Field(..., description="Valor da transação")
    description: Optional[str] = Field(None, description="Descrição da transação")
    status: str = Field(..., description="Status da transação")
    institution: str = Field(..., description="Instituição financeira")
    counterparty: Optional[str] = Field(None, description="Contraparte da transação")
    category: Optional[str] = Field(None, description="Categoria da transação")
    metadata: Dict[str, Any] = Field(default_factory=dict, description="Metadados adicionais")


class TransactionHistory(BaseModel):
    """Histórico de transações"""
    document: str = Field(..., description="Número do documento (CPF/CNPJ)")
    provider: str = Field(..., description="Provedor do bureau de crédito")
    date: datetime = Field(..., description="Data da consulta")
    transactions: List[Transaction] = Field(default_factory=list, description="Lista de transações")
    start_date: Optional[datetime] = Field(None, description="Data inicial do período")
    end_date: Optional[datetime] = Field(None, description="Data final do período")
    total_count: int = Field(..., description="Total de transações")
    total_amount: float = Field(..., description="Valor total das transações")
    metadata: Dict[str, Any] = Field(default_factory=dict, description="Metadados adicionais")


class FinancialProfile(BaseModel):
    """Perfil financeiro do cliente"""
    document: str = Field(..., description="Número do documento (CPF/CNPJ)")
    name: str = Field(..., description="Nome da pessoa ou empresa")
    provider: str = Field(..., description="Provedor do bureau de crédito")
    date: datetime = Field(..., description="Data da consulta")
    income_range: Optional[str] = Field(None, description="Faixa de renda estimada")
    estimated_income: Optional[float] = Field(None, description="Renda estimada")
    employment_status: Optional[str] = Field(None, description="Situação de emprego")
    risk_profile: str = Field(..., description="Perfil de risco")
    segment: str = Field(..., description="Segmento do cliente")
    avg_monthly_expenses: Optional[float] = Field(None, description="Despesas médias mensais")
    avg_monthly_income: Optional[float] = Field(None, description="Receita média mensal")
    financial_maturity: Optional[int] = Field(None, 
                                           description="Maturidade financeira (0-100)")
    suggested_credit_limit: Optional[float] = Field(None, description="Limite de crédito sugerido")
    payment_behavior: Optional[Dict[str, Any]] = Field(None, 
                                                    description="Comportamento de pagamento")
    assets: Optional[List[Dict[str, Any]]] = Field(None, description="Ativos financeiros")
    liabilities: Optional[List[Dict[str, Any]]] = Field(None, description="Passivos financeiros")
    metadata: Dict[str, Any] = Field(default_factory=dict, description="Metadados adicionais")
    
    @validator("financial_maturity")
    def validate_financial_maturity(cls, v):
        if v is not None and not 0 <= v <= 100:
            raise ValueError("Financial maturity must be between 0 and 100")
        return v