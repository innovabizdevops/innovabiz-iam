-- CreateTable
CREATE TABLE "agents" (
    "id" SERIAL NOT NULL,
    "name" VARCHAR(200) NOT NULL,
    "type" VARCHAR(50) NOT NULL,
    "description" TEXT,
    "use_cases" VARCHAR[] DEFAULT ARRAY[]::VARCHAR[],
    "created_at" TIMESTAMP(6) DEFAULT CURRENT_TIMESTAMP,
    "updated_at" TIMESTAMP(6) DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT "agents_pkey" PRIMARY KEY ("id")
);

-- CreateTable
CREATE TABLE "alembic_version" (
    "version_num" VARCHAR(32) NOT NULL,

    CONSTRAINT "alembic_version_pkc" PRIMARY KEY ("version_num")
);

-- CreateTable
CREATE TABLE "capabilities" (
    "id" VARCHAR(20) NOT NULL,
    "name" VARCHAR(500) NOT NULL,
    "level" INTEGER NOT NULL,
    "description" TEXT,
    "parent_id" VARCHAR(20),
    "metadata" JSONB DEFAULT '{}',
    "data_quality_score" DECIMAL(3,2),
    "data_owner" VARCHAR(100),
    "data_steward" VARCHAR(100),
    "tenant_id" VARCHAR(50) NOT NULL DEFAULT 'default',
    "created_at" TIMESTAMP(6) DEFAULT CURRENT_TIMESTAMP,
    "created_by" VARCHAR(100),
    "updated_at" TIMESTAMP(6) DEFAULT CURRENT_TIMESTAMP,
    "updated_by" VARCHAR(100),
    "search_vector" tsvector DEFAULT to_tsvector('english'::regconfig, (((COALESCE(name, ''::character varying))::text || ' '::text) || COALESCE(description, ''::text))),

    CONSTRAINT "capabilities_pkey" PRIMARY KEY ("id")
);

-- CreateTable
CREATE TABLE "capability_geography" (
    "capability_id" VARCHAR(20) NOT NULL,
    "country_code" VARCHAR(2) NOT NULL,
    "tenant_id" VARCHAR(50) NOT NULL DEFAULT 'default',
    "created_at" TIMESTAMP(6) DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT "capability_geography_pkey" PRIMARY KEY ("capability_id","country_code")
);

-- CreateTable
CREATE TABLE "capability_industries" (
    "capability_id" VARCHAR(20) NOT NULL,
    "industry_id" INTEGER NOT NULL,
    "tenant_id" VARCHAR(50) NOT NULL DEFAULT 'default',
    "created_at" TIMESTAMP(6) DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT "capability_industries_pkey" PRIMARY KEY ("capability_id","industry_id")
);

-- CreateTable
CREATE TABLE "capability_modules" (
    "capability_id" VARCHAR(20) NOT NULL,
    "module_id" INTEGER NOT NULL,
    "relationship_type" VARCHAR(50),
    "tenant_id" VARCHAR(50) NOT NULL DEFAULT 'default',
    "created_at" TIMESTAMP(6) DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT "capability_modules_pkey" PRIMARY KEY ("capability_id","module_id")
);

-- CreateTable
CREATE TABLE "capability_organizational" (
    "capability_id" VARCHAR(20) NOT NULL,
    "company_types" VARCHAR[] DEFAULT ARRAY[]::VARCHAR[],
    "org_structures" VARCHAR[] DEFAULT ARRAY[]::VARCHAR[],
    "business_segments" VARCHAR[] DEFAULT ARRAY[]::VARCHAR[],
    "service_types" VARCHAR[] DEFAULT ARRAY[]::VARCHAR[],
    "product_types" VARCHAR[] DEFAULT ARRAY[]::VARCHAR[],
    "customer_types" VARCHAR[] DEFAULT ARRAY[]::VARCHAR[],
    "tenant_id" VARCHAR(50) NOT NULL DEFAULT 'default',

    CONSTRAINT "capability_organizational_pkey" PRIMARY KEY ("capability_id")
);

-- CreateTable
CREATE TABLE "capability_performance" (
    "capability_id" VARCHAR(20) NOT NULL,
    "bsc_perspective" VARCHAR(100),
    "indicator_types" VARCHAR[] DEFAULT ARRAY[]::VARCHAR[],
    "swot_relevance" VARCHAR[] DEFAULT ARRAY[]::VARCHAR[],
    "results_types" VARCHAR[] DEFAULT ARRAY[]::VARCHAR[],
    "tenant_id" VARCHAR(50) NOT NULL DEFAULT 'default',

    CONSTRAINT "capability_performance_pkey" PRIMARY KEY ("capability_id")
);

-- CreateTable
CREATE TABLE "capability_strategy" (
    "capability_id" VARCHAR(20) NOT NULL,
    "relevant_strategies" VARCHAR[] DEFAULT ARRAY[]::VARCHAR[],
    "applicable_market_types" VARCHAR[] DEFAULT ARRAY[]::VARCHAR[],
    "competitor_relevance" VARCHAR[] DEFAULT ARRAY[]::VARCHAR[],
    "plan_types" VARCHAR[] DEFAULT ARRAY[]::VARCHAR[],
    "tenant_id" VARCHAR(50) NOT NULL DEFAULT 'default',

    CONSTRAINT "capability_strategy_pkey" PRIMARY KEY ("capability_id")
);

-- CreateTable
CREATE TABLE "countries" (
    "code" VARCHAR(2) NOT NULL,
    "name" VARCHAR(100) NOT NULL,
    "continent" VARCHAR(50) NOT NULL,
    "region" VARCHAR(100),
    "population" BIGINT,
    "gdp_usd" BIGINT,
    "created_at" TIMESTAMP(6) DEFAULT CURRENT_TIMESTAMP,
    "updated_at" TIMESTAMP(6) DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT "countries_pkey" PRIMARY KEY ("code")
);

-- CreateTable
CREATE TABLE "industries" (
    "id" SERIAL NOT NULL,
    "name" VARCHAR(200) NOT NULL,
    "category" VARCHAR(100) NOT NULL,
    "sub_sectors" VARCHAR[] DEFAULT ARRAY[]::VARCHAR[],
    "description" TEXT,
    "created_at" TIMESTAMP(6) DEFAULT CURRENT_TIMESTAMP,
    "updated_at" TIMESTAMP(6) DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT "industries_pkey" PRIMARY KEY ("id")
);

-- CreateTable
CREATE TABLE "modules" (
    "id" SERIAL NOT NULL,
    "name" VARCHAR(200) NOT NULL,
    "category" VARCHAR(100) NOT NULL,
    "description" TEXT,
    "dependencies" JSONB DEFAULT '{}',
    "created_at" TIMESTAMP(6) DEFAULT CURRENT_TIMESTAMP,
    "updated_at" TIMESTAMP(6) DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT "modules_pkey" PRIMARY KEY ("id")
);

-- CreateTable
CREATE TABLE "UniversalEntity" (
    "id" TEXT NOT NULL,
    "createdAt" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updatedAt" TIMESTAMP(3) NOT NULL,
    "tenantId" TEXT NOT NULL DEFAULT 'default',
    "domain" TEXT NOT NULL,
    "entityType" TEXT NOT NULL,
    "region" TEXT,
    "industry" TEXT,
    "data" JSONB NOT NULL,
    "tags" TEXT[],
    "deleted" BOOLEAN NOT NULL DEFAULT false,

    CONSTRAINT "UniversalEntity_pkey" PRIMARY KEY ("id")
);

-- CreateTable
CREATE TABLE "ChatSession" (
    "id" TEXT NOT NULL,
    "userId" TEXT NOT NULL DEFAULT 'anonymous',
    "agentId" TEXT NOT NULL,
    "createdAt" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updatedAt" TIMESTAMP(3) NOT NULL,

    CONSTRAINT "ChatSession_pkey" PRIMARY KEY ("id")
);

-- CreateTable
CREATE TABLE "ChatMessage" (
    "id" TEXT NOT NULL,
    "sessionId" TEXT NOT NULL,
    "sender" TEXT NOT NULL,
    "text" TEXT NOT NULL,
    "metadata" JSONB,
    "createdAt" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT "ChatMessage_pkey" PRIMARY KEY ("id")
);

-- CreateTable
CREATE TABLE "HealthcarePatient" (
    "id" TEXT NOT NULL,
    "firstName" TEXT NOT NULL,
    "lastName" TEXT NOT NULL,
    "dob" TIMESTAMP(3) NOT NULL,
    "status" TEXT NOT NULL DEFAULT 'Active',
    "lastVisit" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "tenantId" TEXT NOT NULL DEFAULT 'default',
    "createdAt" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updatedAt" TIMESTAMP(3) NOT NULL,

    CONSTRAINT "HealthcarePatient_pkey" PRIMARY KEY ("id")
);

-- CreateTable
CREATE TABLE "HealthcareAppointment" (
    "id" TEXT NOT NULL,
    "patientId" TEXT NOT NULL,
    "date" TIMESTAMP(3) NOT NULL,
    "doctor" TEXT NOT NULL,
    "status" TEXT NOT NULL DEFAULT 'Scheduled',
    "notes" TEXT,
    "tenantId" TEXT NOT NULL DEFAULT 'default',
    "createdAt" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updatedAt" TIMESTAMP(3) NOT NULL,

    CONSTRAINT "HealthcareAppointment_pkey" PRIMARY KEY ("id")
);

-- CreateTable
CREATE TABLE "FinancialAccount" (
    "id" TEXT NOT NULL,
    "name" TEXT NOT NULL,
    "type" TEXT NOT NULL,
    "balance" DECIMAL(19,4) NOT NULL,
    "currency" TEXT NOT NULL DEFAULT 'USD',
    "tenantId" TEXT NOT NULL DEFAULT 'default',
    "createdAt" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updatedAt" TIMESTAMP(3) NOT NULL,

    CONSTRAINT "FinancialAccount_pkey" PRIMARY KEY ("id")
);

-- CreateTable
CREATE TABLE "FinancialTransaction" (
    "id" TEXT NOT NULL,
    "accountId" TEXT NOT NULL,
    "amount" DECIMAL(19,4) NOT NULL,
    "type" TEXT NOT NULL,
    "category" TEXT,
    "status" TEXT NOT NULL DEFAULT 'Pending',
    "date" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "tenantId" TEXT NOT NULL DEFAULT 'default',
    "createdAt" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updatedAt" TIMESTAMP(3) NOT NULL,

    CONSTRAINT "FinancialTransaction_pkey" PRIMARY KEY ("id")
);

-- CreateTable
CREATE TABLE "GovernmentCitizen" (
    "id" TEXT NOT NULL,
    "nationalId" TEXT NOT NULL,
    "firstName" TEXT NOT NULL,
    "lastName" TEXT NOT NULL,
    "address" TEXT,
    "city" TEXT,
    "tenantId" TEXT NOT NULL DEFAULT 'default',
    "createdAt" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updatedAt" TIMESTAMP(3) NOT NULL,

    CONSTRAINT "GovernmentCitizen_pkey" PRIMARY KEY ("id")
);

-- CreateTable
CREATE TABLE "GovernmentServiceRequest" (
    "id" TEXT NOT NULL,
    "citizenId" TEXT NOT NULL,
    "serviceType" TEXT NOT NULL,
    "status" TEXT NOT NULL,
    "tenantId" TEXT NOT NULL DEFAULT 'default',
    "createdAt" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updatedAt" TIMESTAMP(3) NOT NULL,

    CONSTRAINT "GovernmentServiceRequest_pkey" PRIMARY KEY ("id")
);

-- CreateTable
CREATE TABLE "EducationStudent" (
    "id" TEXT NOT NULL,
    "studentId" TEXT NOT NULL,
    "firstName" TEXT NOT NULL,
    "lastName" TEXT NOT NULL,
    "email" TEXT,
    "tenantId" TEXT NOT NULL DEFAULT 'default',
    "createdAt" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updatedAt" TIMESTAMP(3) NOT NULL,

    CONSTRAINT "EducationStudent_pkey" PRIMARY KEY ("id")
);

-- CreateTable
CREATE TABLE "EducationCourse" (
    "id" TEXT NOT NULL,
    "code" TEXT NOT NULL,
    "name" TEXT NOT NULL,
    "credits" INTEGER NOT NULL,
    "tenantId" TEXT NOT NULL DEFAULT 'default',
    "createdAt" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updatedAt" TIMESTAMP(3) NOT NULL,

    CONSTRAINT "EducationCourse_pkey" PRIMARY KEY ("id")
);

-- CreateTable
CREATE TABLE "EducationEnrollment" (
    "id" TEXT NOT NULL,
    "studentId" TEXT NOT NULL,
    "courseId" TEXT NOT NULL,
    "grade" TEXT,
    "tenantId" TEXT NOT NULL DEFAULT 'default',
    "createdAt" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updatedAt" TIMESTAMP(3) NOT NULL,

    CONSTRAINT "EducationEnrollment_pkey" PRIMARY KEY ("id")
);

-- CreateTable
CREATE TABLE "RetailProduct" (
    "id" TEXT NOT NULL,
    "sku" TEXT NOT NULL,
    "name" TEXT NOT NULL,
    "price" DECIMAL(10,2) NOT NULL,
    "stock" INTEGER NOT NULL,
    "category" TEXT,
    "tenantId" TEXT NOT NULL DEFAULT 'default',
    "createdAt" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updatedAt" TIMESTAMP(3) NOT NULL,

    CONSTRAINT "RetailProduct_pkey" PRIMARY KEY ("id")
);

-- CreateTable
CREATE TABLE "RetailOrder" (
    "id" TEXT NOT NULL,
    "customer" TEXT,
    "status" TEXT NOT NULL,
    "total" DECIMAL(10,2) NOT NULL,
    "tenantId" TEXT NOT NULL DEFAULT 'default',
    "createdAt" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updatedAt" TIMESTAMP(3) NOT NULL,

    CONSTRAINT "RetailOrder_pkey" PRIMARY KEY ("id")
);

-- CreateTable
CREATE TABLE "RetailOrderItem" (
    "id" TEXT NOT NULL,
    "orderId" TEXT NOT NULL,
    "productId" TEXT NOT NULL,
    "quantity" INTEGER NOT NULL,
    "price" DECIMAL(10,2) NOT NULL,

    CONSTRAINT "RetailOrderItem_pkey" PRIMARY KEY ("id")
);

-- CreateTable
CREATE TABLE "IamUser" (
    "id" TEXT NOT NULL,
    "email" TEXT NOT NULL,
    "displayName" TEXT NOT NULL,
    "status" TEXT NOT NULL DEFAULT 'ACTIVE',
    "source" TEXT NOT NULL DEFAULT 'LOCAL',
    "federationProvider" TEXT,
    "region" TEXT NOT NULL DEFAULT 'EU-PT',
    "jurisdictions" TEXT[] DEFAULT ARRAY[]::TEXT[],
    "consentTimestamp" TIMESTAMP(3),
    "dataResidency" TEXT,
    "mfaEnrolled" BOOLEAN NOT NULL DEFAULT false,
    "riskScore" INTEGER NOT NULL DEFAULT 0,
    "sovereign" JSONB,
    "cognitive" JSONB,
    "metadata" JSONB,
    "provisionedAt" TIMESTAMP(3),
    "lastLoginAt" TIMESTAMP(3),
    "passwordChangedAt" TIMESTAMP(3),
    "tenantId" TEXT NOT NULL DEFAULT 'default',
    "createdAt" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updatedAt" TIMESTAMP(3) NOT NULL,

    CONSTRAINT "IamUser_pkey" PRIMARY KEY ("id")
);

-- CreateTable
CREATE TABLE "IamRole" (
    "id" TEXT NOT NULL,
    "name" TEXT NOT NULL,
    "description" TEXT,
    "authzModel" TEXT NOT NULL DEFAULT 'RBAC',
    "permissions" TEXT[] DEFAULT ARRAY[]::TEXT[],
    "riskLevel" TEXT NOT NULL DEFAULT 'low',
    "hierarchy" JSONB,
    "constraints" JSONB,
    "metadata" JSONB,
    "tenantId" TEXT NOT NULL DEFAULT 'default',
    "createdAt" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updatedAt" TIMESTAMP(3) NOT NULL,

    CONSTRAINT "IamRole_pkey" PRIMARY KEY ("id")
);

-- CreateTable
CREATE TABLE "IamUserRole" (
    "id" TEXT NOT NULL,
    "userId" TEXT NOT NULL,
    "roleId" TEXT NOT NULL,
    "grantedBy" TEXT,
    "expiresAt" TIMESTAMP(3),
    "tenantId" TEXT NOT NULL DEFAULT 'default',
    "createdAt" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT "IamUserRole_pkey" PRIMARY KEY ("id")
);

-- CreateTable
CREATE TABLE "IamSession" (
    "id" TEXT NOT NULL,
    "userId" TEXT NOT NULL,
    "isActive" BOOLEAN NOT NULL DEFAULT true,
    "trustScore" INTEGER NOT NULL DEFAULT 50,
    "trustLevel" TEXT NOT NULL DEFAULT 'MEDIUM',
    "mfaLevel" TEXT NOT NULL DEFAULT 'AAL1',
    "deviceFingerprint" TEXT,
    "geoLocation" JSONB,
    "ipAddress" TEXT,
    "userAgent" TEXT,
    "terminatedAt" TIMESTAMP(3),
    "terminationReason" TEXT,
    "tenantId" TEXT NOT NULL DEFAULT 'default',
    "startedAt" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updatedAt" TIMESTAMP(3) NOT NULL,

    CONSTRAINT "IamSession_pkey" PRIMARY KEY ("id")
);

-- CreateTable
CREATE TABLE "IamMfaDevice" (
    "id" TEXT NOT NULL,
    "userId" TEXT NOT NULL,
    "type" TEXT NOT NULL,
    "aalLevel" TEXT NOT NULL DEFAULT 'AAL2',
    "displayName" TEXT,
    "isVerified" BOOLEAN NOT NULL DEFAULT false,
    "phishingSafe" BOOLEAN NOT NULL DEFAULT false,
    "credentialId" TEXT,
    "lastUsedAt" TIMESTAMP(3),
    "tenantId" TEXT NOT NULL DEFAULT 'default',
    "registeredAt" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updatedAt" TIMESTAMP(3) NOT NULL,

    CONSTRAINT "IamMfaDevice_pkey" PRIMARY KEY ("id")
);

-- CreateTable
CREATE TABLE "IamApiKey" (
    "id" TEXT NOT NULL,
    "name" TEXT NOT NULL,
    "type" TEXT NOT NULL DEFAULT 'SERVICE',
    "keyHash" TEXT,
    "scopes" TEXT[] DEFAULT ARRAY[]::TEXT[],
    "isActive" BOOLEAN NOT NULL DEFAULT true,
    "rateLimitPerMin" INTEGER NOT NULL DEFAULT 1000,
    "lastUsedAt" TIMESTAMP(3),
    "expiresAt" TIMESTAMP(3),
    "revokedAt" TIMESTAMP(3),
    "revokedBy" TEXT,
    "userId" TEXT,
    "tenantId" TEXT NOT NULL DEFAULT 'default',
    "createdAt" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updatedAt" TIMESTAMP(3) NOT NULL,

    CONSTRAINT "IamApiKey_pkey" PRIMARY KEY ("id")
);

-- CreateTable
CREATE TABLE "IamPrivilegedAccount" (
    "id" TEXT NOT NULL,
    "accountName" TEXT NOT NULL,
    "accountType" TEXT NOT NULL,
    "privilegeLevel" TEXT NOT NULL DEFAULT 'STANDARD',
    "vaultStatus" TEXT NOT NULL DEFAULT 'VAULTED',
    "jitEnabled" BOOLEAN NOT NULL DEFAULT false,
    "sessionRecording" BOOLEAN NOT NULL DEFAULT true,
    "lastRotation" TIMESTAMP(3),
    "nextRotation" TIMESTAMP(3),
    "rotationPolicy" JSONB,
    "lastCheckout" JSONB,
    "tenantId" TEXT NOT NULL DEFAULT 'default',
    "createdAt" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updatedAt" TIMESTAMP(3) NOT NULL,

    CONSTRAINT "IamPrivilegedAccount_pkey" PRIMARY KEY ("id")
);

-- CreateTable
CREATE TABLE "GrcFramework" (
    "id" TEXT NOT NULL,
    "name" TEXT NOT NULL,
    "type" TEXT NOT NULL,
    "version" TEXT,
    "status" TEXT NOT NULL DEFAULT 'DRAFT',
    "scope" TEXT[] DEFAULT ARRAY[]::TEXT[],
    "controls" JSONB,
    "riskAppetite" JSONB,
    "metadata" JSONB,
    "tenantId" TEXT NOT NULL DEFAULT 'default',
    "createdAt" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updatedAt" TIMESTAMP(3) NOT NULL,
    "createdBy" TEXT,
    "updatedBy" TEXT,

    CONSTRAINT "GrcFramework_pkey" PRIMARY KEY ("id")
);

-- CreateTable
CREATE TABLE "ComplianceRequirement" (
    "id" TEXT NOT NULL,
    "regulation" TEXT NOT NULL,
    "article" TEXT,
    "description" TEXT NOT NULL,
    "status" TEXT NOT NULL DEFAULT 'PENDING',
    "dueDate" TIMESTAMP(3),
    "owner" TEXT,
    "evidence" JSONB,
    "riskLevel" TEXT NOT NULL DEFAULT 'medium',
    "metadata" JSONB,
    "tenantId" TEXT NOT NULL DEFAULT 'default',
    "createdAt" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updatedAt" TIMESTAMP(3) NOT NULL,
    "createdBy" TEXT,
    "updatedBy" TEXT,

    CONSTRAINT "ComplianceRequirement_pkey" PRIMARY KEY ("id")
);

-- CreateTable
CREATE TABLE "AuditFinding" (
    "id" TEXT NOT NULL,
    "auditId" TEXT NOT NULL,
    "title" TEXT NOT NULL,
    "type" TEXT NOT NULL,
    "status" TEXT NOT NULL DEFAULT 'OPEN',
    "description" TEXT NOT NULL,
    "rootCause" TEXT,
    "recommendation" TEXT,
    "assignee" TEXT,
    "dueDate" TIMESTAMP(3),
    "closedAt" TIMESTAMP(3),
    "evidence" JSONB,
    "metadata" JSONB,
    "tenantId" TEXT NOT NULL DEFAULT 'default',
    "createdAt" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updatedAt" TIMESTAMP(3) NOT NULL,
    "createdBy" TEXT,
    "updatedBy" TEXT,

    CONSTRAINT "AuditFinding_pkey" PRIMARY KEY ("id")
);

-- CreateTable
CREATE TABLE "RiskRegister" (
    "id" TEXT NOT NULL,
    "title" TEXT NOT NULL,
    "category" TEXT NOT NULL,
    "likelihood" INTEGER NOT NULL,
    "impact" INTEGER NOT NULL,
    "riskScore" INTEGER,
    "status" TEXT NOT NULL DEFAULT 'IDENTIFIED',
    "owner" TEXT,
    "mitigations" JSONB,
    "description" TEXT,
    "residualRisk" INTEGER,
    "metadata" JSONB,
    "tenantId" TEXT NOT NULL DEFAULT 'default',
    "createdAt" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updatedAt" TIMESTAMP(3) NOT NULL,
    "createdBy" TEXT,
    "updatedBy" TEXT,

    CONSTRAINT "RiskRegister_pkey" PRIMARY KEY ("id")
);

-- CreateTable
CREATE TABLE "Contract" (
    "id" TEXT NOT NULL,
    "title" TEXT NOT NULL,
    "type" TEXT NOT NULL,
    "status" TEXT NOT NULL DEFAULT 'DRAFT',
    "counterparty" TEXT NOT NULL,
    "value" DECIMAL(19,4),
    "currency" TEXT NOT NULL DEFAULT 'USD',
    "startDate" TIMESTAMP(3),
    "endDate" TIMESTAMP(3),
    "renewalType" TEXT,
    "terms" JSONB,
    "signatories" JSONB,
    "metadata" JSONB,
    "tenantId" TEXT NOT NULL DEFAULT 'default',
    "createdAt" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updatedAt" TIMESTAMP(3) NOT NULL,
    "createdBy" TEXT,
    "updatedBy" TEXT,

    CONSTRAINT "Contract_pkey" PRIMARY KEY ("id")
);

-- CreateTable
CREATE TABLE "HcmEmployee" (
    "id" TEXT NOT NULL,
    "employeeId" TEXT NOT NULL,
    "firstName" TEXT NOT NULL,
    "lastName" TEXT NOT NULL,
    "email" TEXT NOT NULL,
    "department" TEXT,
    "position" TEXT,
    "level" TEXT,
    "status" TEXT NOT NULL DEFAULT 'ACTIVE',
    "hireDate" TIMESTAMP(3),
    "manager" TEXT,
    "compensation" JSONB,
    "skills" TEXT[] DEFAULT ARRAY[]::TEXT[],
    "certifications" TEXT[] DEFAULT ARRAY[]::TEXT[],
    "metadata" JSONB,
    "tenantId" TEXT NOT NULL DEFAULT 'default',
    "createdAt" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updatedAt" TIMESTAMP(3) NOT NULL,
    "createdBy" TEXT,
    "updatedBy" TEXT,

    CONSTRAINT "HcmEmployee_pkey" PRIMARY KEY ("id")
);

-- CreateTable
CREATE TABLE "MarketingCampaign" (
    "id" TEXT NOT NULL,
    "name" TEXT NOT NULL,
    "type" TEXT NOT NULL,
    "status" TEXT NOT NULL DEFAULT 'DRAFT',
    "channel" TEXT[] DEFAULT ARRAY[]::TEXT[],
    "budget" DECIMAL(19,4),
    "spent" DECIMAL(19,4),
    "currency" TEXT NOT NULL DEFAULT 'USD',
    "startDate" TIMESTAMP(3),
    "endDate" TIMESTAMP(3),
    "audience" JSONB,
    "metrics" JSONB,
    "content" JSONB,
    "metadata" JSONB,
    "tenantId" TEXT NOT NULL DEFAULT 'default',
    "createdAt" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updatedAt" TIMESTAMP(3) NOT NULL,
    "createdBy" TEXT,
    "updatedBy" TEXT,

    CONSTRAINT "MarketingCampaign_pkey" PRIMARY KEY ("id")
);

-- CreateTable
CREATE TABLE "Partner" (
    "id" TEXT NOT NULL,
    "name" TEXT NOT NULL,
    "type" TEXT NOT NULL,
    "status" TEXT NOT NULL DEFAULT 'PROSPECT',
    "tier" TEXT,
    "contactEmail" TEXT,
    "contactPhone" TEXT,
    "region" TEXT,
    "industry" TEXT,
    "revenue" DECIMAL(19,4),
    "agreement" JSONB,
    "certifications" TEXT[] DEFAULT ARRAY[]::TEXT[],
    "metadata" JSONB,
    "tenantId" TEXT NOT NULL DEFAULT 'default',
    "createdAt" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updatedAt" TIMESTAMP(3) NOT NULL,
    "createdBy" TEXT,
    "updatedBy" TEXT,

    CONSTRAINT "Partner_pkey" PRIMARY KEY ("id")
);

-- CreateTable
CREATE TABLE "ProcessWorkflow" (
    "id" TEXT NOT NULL,
    "name" TEXT NOT NULL,
    "category" TEXT NOT NULL,
    "status" TEXT NOT NULL DEFAULT 'DRAFT',
    "version" INTEGER NOT NULL DEFAULT 1,
    "steps" JSONB NOT NULL,
    "triggers" JSONB,
    "sla" JSONB,
    "metadata" JSONB,
    "tenantId" TEXT NOT NULL DEFAULT 'default',
    "createdAt" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updatedAt" TIMESTAMP(3) NOT NULL,
    "createdBy" TEXT,
    "updatedBy" TEXT,

    CONSTRAINT "ProcessWorkflow_pkey" PRIMARY KEY ("id")
);

-- CreateTable
CREATE TABLE "QualityCheck" (
    "id" TEXT NOT NULL,
    "name" TEXT NOT NULL,
    "type" TEXT NOT NULL,
    "status" TEXT NOT NULL DEFAULT 'PENDING',
    "standard" TEXT,
    "targetEntity" TEXT,
    "checklist" JSONB,
    "score" DECIMAL(5,2),
    "inspector" TEXT,
    "completedAt" TIMESTAMP(3),
    "metadata" JSONB,
    "tenantId" TEXT NOT NULL DEFAULT 'default',
    "createdAt" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updatedAt" TIMESTAMP(3) NOT NULL,
    "createdBy" TEXT,
    "updatedBy" TEXT,

    CONSTRAINT "QualityCheck_pkey" PRIMARY KEY ("id")
);

-- CreateTable
CREATE TABLE "KnowledgeArticle" (
    "id" TEXT NOT NULL,
    "title" TEXT NOT NULL,
    "category" TEXT NOT NULL,
    "status" TEXT NOT NULL DEFAULT 'DRAFT',
    "content" TEXT NOT NULL,
    "summary" TEXT,
    "author" TEXT,
    "tags" TEXT[] DEFAULT ARRAY[]::TEXT[],
    "views" INTEGER NOT NULL DEFAULT 0,
    "rating" DECIMAL(3,2),
    "version" INTEGER NOT NULL DEFAULT 1,
    "language" TEXT NOT NULL DEFAULT 'en',
    "metadata" JSONB,
    "tenantId" TEXT NOT NULL DEFAULT 'default',
    "createdAt" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updatedAt" TIMESTAMP(3) NOT NULL,
    "createdBy" TEXT,
    "updatedBy" TEXT,

    CONSTRAINT "KnowledgeArticle_pkey" PRIMARY KEY ("id")
);

-- CreateTable
CREATE TABLE "VendorRecord" (
    "id" TEXT NOT NULL,
    "name" TEXT NOT NULL,
    "category" TEXT NOT NULL,
    "status" TEXT NOT NULL DEFAULT 'ACTIVE',
    "riskRating" TEXT,
    "contactEmail" TEXT,
    "country" TEXT,
    "currency" TEXT NOT NULL DEFAULT 'USD',
    "paymentTerms" TEXT,
    "annualSpend" DECIMAL(19,4),
    "compliance" JSONB,
    "performance" JSONB,
    "metadata" JSONB,
    "tenantId" TEXT NOT NULL DEFAULT 'default',
    "createdAt" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updatedAt" TIMESTAMP(3) NOT NULL,
    "createdBy" TEXT,
    "updatedBy" TEXT,

    CONSTRAINT "VendorRecord_pkey" PRIMARY KEY ("id")
);

-- CreateTable
CREATE TABLE "DataAsset" (
    "id" TEXT NOT NULL,
    "name" TEXT NOT NULL,
    "type" TEXT NOT NULL,
    "classification" TEXT NOT NULL DEFAULT 'INTERNAL',
    "status" TEXT NOT NULL DEFAULT 'ACTIVE',
    "owner" TEXT,
    "steward" TEXT,
    "domain" TEXT,
    "schema" JSONB,
    "lineage" JSONB,
    "quality" JSONB,
    "metadata" JSONB,
    "tenantId" TEXT NOT NULL DEFAULT 'default',
    "createdAt" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updatedAt" TIMESTAMP(3) NOT NULL,
    "createdBy" TEXT,
    "updatedBy" TEXT,

    CONSTRAINT "DataAsset_pkey" PRIMARY KEY ("id")
);

-- CreateTable
CREATE TABLE "Device" (
    "id" TEXT NOT NULL,
    "name" TEXT NOT NULL,
    "type" TEXT NOT NULL,
    "status" TEXT NOT NULL DEFAULT 'ACTIVE',
    "serialNumber" TEXT,
    "manufacturer" TEXT,
    "model" TEXT,
    "firmware" TEXT,
    "ipAddress" TEXT,
    "location" TEXT,
    "lastSeen" TIMESTAMP(3),
    "compliance" JSONB,
    "telemetry" JSONB,
    "metadata" JSONB,
    "tenantId" TEXT NOT NULL DEFAULT 'default',
    "createdAt" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updatedAt" TIMESTAMP(3) NOT NULL,
    "createdBy" TEXT,
    "updatedBy" TEXT,

    CONSTRAINT "Device_pkey" PRIMARY KEY ("id")
);

-- CreateTable
CREATE TABLE "IntegrationConfig" (
    "id" TEXT NOT NULL,
    "name" TEXT NOT NULL,
    "type" TEXT NOT NULL,
    "provider" TEXT,
    "status" TEXT NOT NULL DEFAULT 'ACTIVE',
    "direction" TEXT NOT NULL DEFAULT 'BIDIRECTIONAL',
    "endpoint" TEXT,
    "authType" TEXT,
    "schedule" TEXT,
    "mappings" JSONB,
    "health" JSONB,
    "retryPolicy" JSONB,
    "metadata" JSONB,
    "tenantId" TEXT NOT NULL DEFAULT 'default',
    "createdAt" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updatedAt" TIMESTAMP(3) NOT NULL,
    "createdBy" TEXT,
    "updatedBy" TEXT,

    CONSTRAINT "IntegrationConfig_pkey" PRIMARY KEY ("id")
);

-- CreateTable
CREATE TABLE "SupportTicket" (
    "id" TEXT NOT NULL,
    "ticketNumber" TEXT NOT NULL,
    "subject" TEXT NOT NULL,
    "channel" TEXT NOT NULL,
    "priority" TEXT NOT NULL DEFAULT 'MEDIUM',
    "status" TEXT NOT NULL DEFAULT 'NEW',
    "category" TEXT,
    "description" TEXT NOT NULL,
    "assignee" TEXT,
    "requester" TEXT,
    "firstResponse" TIMESTAMP(3),
    "resolvedAt" TIMESTAMP(3),
    "satisfaction" INTEGER,
    "sla" JSONB,
    "metadata" JSONB,
    "tenantId" TEXT NOT NULL DEFAULT 'default',
    "createdAt" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updatedAt" TIMESTAMP(3) NOT NULL,
    "createdBy" TEXT,
    "updatedBy" TEXT,

    CONSTRAINT "SupportTicket_pkey" PRIMARY KEY ("id")
);

-- CreateTable
CREATE TABLE "DocumentRecord" (
    "id" TEXT NOT NULL,
    "title" TEXT NOT NULL,
    "type" TEXT NOT NULL,
    "format" TEXT,
    "status" TEXT NOT NULL DEFAULT 'DRAFT',
    "version" INTEGER NOT NULL DEFAULT 1,
    "storageUrl" TEXT,
    "size" BIGINT,
    "author" TEXT,
    "tags" TEXT[] DEFAULT ARRAY[]::TEXT[],
    "permissions" JSONB,
    "retention" JSONB,
    "metadata" JSONB,
    "tenantId" TEXT NOT NULL DEFAULT 'default',
    "createdAt" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updatedAt" TIMESTAMP(3) NOT NULL,
    "createdBy" TEXT,
    "updatedBy" TEXT,

    CONSTRAINT "DocumentRecord_pkey" PRIMARY KEY ("id")
);

-- CreateTable
CREATE TABLE "NotificationTemplate" (
    "id" TEXT NOT NULL,
    "name" TEXT NOT NULL,
    "channel" TEXT NOT NULL,
    "status" TEXT NOT NULL DEFAULT 'ACTIVE',
    "subject" TEXT,
    "body" TEXT NOT NULL,
    "bodyHtml" TEXT,
    "variables" TEXT[] DEFAULT ARRAY[]::TEXT[],
    "locale" TEXT NOT NULL DEFAULT 'en',
    "priority" TEXT NOT NULL DEFAULT 'NORMAL',
    "metadata" JSONB,
    "tenantId" TEXT NOT NULL DEFAULT 'default',
    "createdAt" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updatedAt" TIMESTAMP(3) NOT NULL,
    "createdBy" TEXT,
    "updatedBy" TEXT,

    CONSTRAINT "NotificationTemplate_pkey" PRIMARY KEY ("id")
);

-- CreateTable
CREATE TABLE "OpenBankingApi" (
    "id" TEXT NOT NULL,
    "name" TEXT NOT NULL,
    "version" TEXT NOT NULL,
    "standard" TEXT NOT NULL,
    "status" TEXT NOT NULL DEFAULT 'ACTIVE',
    "endpoint" TEXT NOT NULL,
    "scope" TEXT[] DEFAULT ARRAY[]::TEXT[],
    "rateLimit" INTEGER,
    "provider" TEXT,
    "certification" JSONB,
    "metadata" JSONB,
    "tenantId" TEXT NOT NULL DEFAULT 'default',
    "createdAt" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updatedAt" TIMESTAMP(3) NOT NULL,

    CONSTRAINT "OpenBankingApi_pkey" PRIMARY KEY ("id")
);

-- CreateTable
CREATE TABLE "OpenFinanceProduct" (
    "id" TEXT NOT NULL,
    "name" TEXT NOT NULL,
    "category" TEXT NOT NULL,
    "provider" TEXT NOT NULL,
    "status" TEXT NOT NULL DEFAULT 'ACTIVE',
    "riskLevel" TEXT,
    "returns" JSONB,
    "fees" JSONB,
    "regulation" TEXT,
    "metadata" JSONB,
    "tenantId" TEXT NOT NULL DEFAULT 'default',
    "createdAt" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updatedAt" TIMESTAMP(3) NOT NULL,

    CONSTRAINT "OpenFinanceProduct_pkey" PRIMARY KEY ("id")
);

-- CreateTable
CREATE TABLE "OpenInsurancePolicy" (
    "id" TEXT NOT NULL,
    "policyNumber" TEXT NOT NULL,
    "type" TEXT NOT NULL,
    "status" TEXT NOT NULL DEFAULT 'ACTIVE',
    "insurer" TEXT NOT NULL,
    "premium" DECIMAL(19,4),
    "currency" TEXT NOT NULL DEFAULT 'USD',
    "coverage" JSONB,
    "startDate" TIMESTAMP(3),
    "endDate" TIMESTAMP(3),
    "holder" JSONB,
    "metadata" JSONB,
    "tenantId" TEXT NOT NULL DEFAULT 'default',
    "createdAt" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updatedAt" TIMESTAMP(3) NOT NULL,

    CONSTRAINT "OpenInsurancePolicy_pkey" PRIMARY KEY ("id")
);

-- CreateTable
CREATE TABLE "OpenDataDataset" (
    "id" TEXT NOT NULL,
    "name" TEXT NOT NULL,
    "category" TEXT NOT NULL,
    "format" TEXT NOT NULL,
    "license" TEXT,
    "status" TEXT NOT NULL DEFAULT 'PUBLISHED',
    "source" TEXT,
    "description" TEXT,
    "recordCount" BIGINT,
    "updateFreq" TEXT,
    "accessUrl" TEXT,
    "quality" JSONB,
    "metadata" JSONB,
    "tenantId" TEXT NOT NULL DEFAULT 'default',
    "createdAt" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updatedAt" TIMESTAMP(3) NOT NULL,

    CONSTRAINT "OpenDataDataset_pkey" PRIMARY KEY ("id")
);

-- CreateTable
CREATE TABLE "OpenHealthResource" (
    "id" TEXT NOT NULL,
    "resourceType" TEXT NOT NULL,
    "fhirVersion" TEXT NOT NULL DEFAULT 'R4',
    "status" TEXT NOT NULL DEFAULT 'ACTIVE',
    "identifier" TEXT,
    "subject" TEXT,
    "data" JSONB NOT NULL,
    "source" TEXT,
    "interop" JSONB,
    "metadata" JSONB,
    "tenantId" TEXT NOT NULL DEFAULT 'default',
    "createdAt" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updatedAt" TIMESTAMP(3) NOT NULL,

    CONSTRAINT "OpenHealthResource_pkey" PRIMARY KEY ("id")
);

-- CreateTable
CREATE TABLE "OpenEducationCourse" (
    "id" TEXT NOT NULL,
    "title" TEXT NOT NULL,
    "provider" TEXT NOT NULL,
    "format" TEXT NOT NULL,
    "standard" TEXT,
    "status" TEXT NOT NULL DEFAULT 'PUBLISHED',
    "level" TEXT,
    "duration" INTEGER,
    "language" TEXT NOT NULL DEFAULT 'en',
    "credits" INTEGER,
    "syllabus" JSONB,
    "pricing" JSONB,
    "metadata" JSONB,
    "tenantId" TEXT NOT NULL DEFAULT 'default',
    "createdAt" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updatedAt" TIMESTAMP(3) NOT NULL,

    CONSTRAINT "OpenEducationCourse_pkey" PRIMARY KEY ("id")
);

-- CreateTable
CREATE TABLE "OpenInnovationChallenge" (
    "id" TEXT NOT NULL,
    "title" TEXT NOT NULL,
    "type" TEXT NOT NULL,
    "status" TEXT NOT NULL DEFAULT 'OPEN',
    "description" TEXT NOT NULL,
    "sponsor" TEXT,
    "prize" JSONB,
    "criteria" JSONB,
    "startDate" TIMESTAMP(3),
    "endDate" TIMESTAMP(3),
    "submissions" INTEGER NOT NULL DEFAULT 0,
    "metadata" JSONB,
    "tenantId" TEXT NOT NULL DEFAULT 'default',
    "createdAt" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updatedAt" TIMESTAMP(3) NOT NULL,

    CONSTRAINT "OpenInnovationChallenge_pkey" PRIMARY KEY ("id")
);

-- CreateTable
CREATE TABLE "InnovationProject" (
    "id" TEXT NOT NULL,
    "name" TEXT NOT NULL,
    "type" TEXT NOT NULL,
    "status" TEXT NOT NULL DEFAULT 'IDEATION',
    "stage" TEXT,
    "owner" TEXT,
    "team" TEXT[] DEFAULT ARRAY[]::TEXT[],
    "budget" DECIMAL(19,4),
    "spent" DECIMAL(19,4),
    "currency" TEXT NOT NULL DEFAULT 'USD',
    "startDate" TIMESTAMP(3),
    "targetDate" TIMESTAMP(3),
    "objectives" JSONB,
    "outcomes" JSONB,
    "metadata" JSONB,
    "tenantId" TEXT NOT NULL DEFAULT 'default',
    "createdAt" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updatedAt" TIMESTAMP(3) NOT NULL,
    "createdBy" TEXT,
    "updatedBy" TEXT,

    CONSTRAINT "InnovationProject_pkey" PRIMARY KEY ("id")
);

-- CreateIndex
CREATE UNIQUE INDEX "agents_name_key" ON "agents"("name");

-- CreateIndex
CREATE INDEX "idx_agents_type" ON "agents"("type");

-- CreateIndex
CREATE INDEX "idx_cap_level_tenant" ON "capabilities"("level", "tenant_id");

-- CreateIndex
CREATE INDEX "idx_cap_metadata" ON "capabilities" USING GIN ("metadata");

-- CreateIndex
CREATE INDEX "idx_cap_name" ON "capabilities"("name");

-- CreateIndex
CREATE INDEX "idx_cap_name_trgm" ON "capabilities" USING GIN ("name" gin_trgm_ops);

-- CreateIndex
CREATE INDEX "idx_cap_search" ON "capabilities" USING GIN ("search_vector");

-- CreateIndex
CREATE INDEX "idx_capgeo_country" ON "capability_geography"("country_code");

-- CreateIndex
CREATE INDEX "idx_capind_industry" ON "capability_industries"("industry_id");

-- CreateIndex
CREATE INDEX "idx_capmod_module" ON "capability_modules"("module_id");

-- CreateIndex
CREATE INDEX "idx_capmod_type" ON "capability_modules"("relationship_type");

-- CreateIndex
CREATE INDEX "idx_org_company_types" ON "capability_organizational" USING GIN ("company_types");

-- CreateIndex
CREATE INDEX "idx_org_segments" ON "capability_organizational" USING GIN ("business_segments");

-- CreateIndex
CREATE INDEX "idx_perf_bsc" ON "capability_performance"("bsc_perspective", "tenant_id");

-- CreateIndex
CREATE INDEX "idx_countries_continent" ON "countries"("continent");

-- CreateIndex
CREATE UNIQUE INDEX "industries_name_key" ON "industries"("name");

-- CreateIndex
CREATE INDEX "idx_industries_category" ON "industries"("category");

-- CreateIndex
CREATE INDEX "idx_industries_sub_sectors" ON "industries" USING GIN ("sub_sectors");

-- CreateIndex
CREATE UNIQUE INDEX "modules_name_key" ON "modules"("name");

-- CreateIndex
CREATE INDEX "idx_modules_category" ON "modules"("category");

-- CreateIndex
CREATE INDEX "idx_modules_dependencies" ON "modules" USING GIN ("dependencies");

-- CreateIndex
CREATE INDEX "UniversalEntity_tenantId_domain_entityType_idx" ON "UniversalEntity"("tenantId", "domain", "entityType");

-- CreateIndex
CREATE INDEX "ChatSession_userId_idx" ON "ChatSession"("userId");

-- CreateIndex
CREATE INDEX "ChatSession_agentId_idx" ON "ChatSession"("agentId");

-- CreateIndex
CREATE INDEX "ChatMessage_sessionId_idx" ON "ChatMessage"("sessionId");

-- CreateIndex
CREATE INDEX "ChatMessage_createdAt_idx" ON "ChatMessage"("createdAt");

-- CreateIndex
CREATE INDEX "HealthcarePatient_tenantId_idx" ON "HealthcarePatient"("tenantId");

-- CreateIndex
CREATE INDEX "HealthcareAppointment_patientId_idx" ON "HealthcareAppointment"("patientId");

-- CreateIndex
CREATE INDEX "HealthcareAppointment_tenantId_idx" ON "HealthcareAppointment"("tenantId");

-- CreateIndex
CREATE INDEX "FinancialAccount_tenantId_idx" ON "FinancialAccount"("tenantId");

-- CreateIndex
CREATE INDEX "FinancialTransaction_accountId_idx" ON "FinancialTransaction"("accountId");

-- CreateIndex
CREATE UNIQUE INDEX "GovernmentCitizen_nationalId_key" ON "GovernmentCitizen"("nationalId");

-- CreateIndex
CREATE INDEX "GovernmentCitizen_tenantId_idx" ON "GovernmentCitizen"("tenantId");

-- CreateIndex
CREATE INDEX "GovernmentServiceRequest_citizenId_idx" ON "GovernmentServiceRequest"("citizenId");

-- CreateIndex
CREATE UNIQUE INDEX "EducationStudent_studentId_key" ON "EducationStudent"("studentId");

-- CreateIndex
CREATE INDEX "EducationStudent_tenantId_idx" ON "EducationStudent"("tenantId");

-- CreateIndex
CREATE UNIQUE INDEX "EducationCourse_code_key" ON "EducationCourse"("code");

-- CreateIndex
CREATE UNIQUE INDEX "EducationEnrollment_studentId_courseId_key" ON "EducationEnrollment"("studentId", "courseId");

-- CreateIndex
CREATE UNIQUE INDEX "RetailProduct_sku_key" ON "RetailProduct"("sku");

-- CreateIndex
CREATE INDEX "RetailProduct_tenantId_idx" ON "RetailProduct"("tenantId");

-- CreateIndex
CREATE INDEX "RetailOrder_tenantId_idx" ON "RetailOrder"("tenantId");

-- CreateIndex
CREATE INDEX "RetailOrderItem_orderId_idx" ON "RetailOrderItem"("orderId");

-- CreateIndex
CREATE INDEX "RetailOrderItem_productId_idx" ON "RetailOrderItem"("productId");

-- CreateIndex
CREATE INDEX "IamUser_tenantId_status_idx" ON "IamUser"("tenantId", "status");

-- CreateIndex
CREATE INDEX "IamUser_tenantId_source_idx" ON "IamUser"("tenantId", "source");

-- CreateIndex
CREATE INDEX "IamUser_region_idx" ON "IamUser"("region");

-- CreateIndex
CREATE UNIQUE INDEX "IamUser_tenantId_email_key" ON "IamUser"("tenantId", "email");

-- CreateIndex
CREATE INDEX "IamRole_tenantId_authzModel_idx" ON "IamRole"("tenantId", "authzModel");

-- CreateIndex
CREATE INDEX "IamRole_riskLevel_idx" ON "IamRole"("riskLevel");

-- CreateIndex
CREATE UNIQUE INDEX "IamRole_tenantId_name_key" ON "IamRole"("tenantId", "name");

-- CreateIndex
CREATE INDEX "IamUserRole_tenantId_idx" ON "IamUserRole"("tenantId");

-- CreateIndex
CREATE UNIQUE INDEX "IamUserRole_userId_roleId_key" ON "IamUserRole"("userId", "roleId");

-- CreateIndex
CREATE INDEX "IamSession_tenantId_userId_idx" ON "IamSession"("tenantId", "userId");

-- CreateIndex
CREATE INDEX "IamSession_tenantId_isActive_idx" ON "IamSession"("tenantId", "isActive");

-- CreateIndex
CREATE INDEX "IamSession_startedAt_idx" ON "IamSession"("startedAt");

-- CreateIndex
CREATE INDEX "IamMfaDevice_tenantId_userId_idx" ON "IamMfaDevice"("tenantId", "userId");

-- CreateIndex
CREATE INDEX "IamMfaDevice_type_idx" ON "IamMfaDevice"("type");

-- CreateIndex
CREATE INDEX "IamApiKey_tenantId_isActive_idx" ON "IamApiKey"("tenantId", "isActive");

-- CreateIndex
CREATE INDEX "IamApiKey_type_idx" ON "IamApiKey"("type");

-- CreateIndex
CREATE UNIQUE INDEX "IamApiKey_tenantId_name_key" ON "IamApiKey"("tenantId", "name");

-- CreateIndex
CREATE INDEX "IamPrivilegedAccount_tenantId_vaultStatus_idx" ON "IamPrivilegedAccount"("tenantId", "vaultStatus");

-- CreateIndex
CREATE INDEX "IamPrivilegedAccount_privilegeLevel_idx" ON "IamPrivilegedAccount"("privilegeLevel");

-- CreateIndex
CREATE INDEX "IamPrivilegedAccount_nextRotation_idx" ON "IamPrivilegedAccount"("nextRotation");

-- CreateIndex
CREATE UNIQUE INDEX "IamPrivilegedAccount_tenantId_accountName_key" ON "IamPrivilegedAccount"("tenantId", "accountName");

-- CreateIndex
CREATE INDEX "GrcFramework_tenantId_type_idx" ON "GrcFramework"("tenantId", "type");

-- CreateIndex
CREATE INDEX "GrcFramework_status_idx" ON "GrcFramework"("status");

-- CreateIndex
CREATE UNIQUE INDEX "GrcFramework_tenantId_name_version_key" ON "GrcFramework"("tenantId", "name", "version");

-- CreateIndex
CREATE INDEX "ComplianceRequirement_tenantId_regulation_idx" ON "ComplianceRequirement"("tenantId", "regulation");

-- CreateIndex
CREATE INDEX "ComplianceRequirement_tenantId_status_idx" ON "ComplianceRequirement"("tenantId", "status");

-- CreateIndex
CREATE INDEX "ComplianceRequirement_dueDate_idx" ON "ComplianceRequirement"("dueDate");

-- CreateIndex
CREATE INDEX "AuditFinding_tenantId_auditId_idx" ON "AuditFinding"("tenantId", "auditId");

-- CreateIndex
CREATE INDEX "AuditFinding_tenantId_status_idx" ON "AuditFinding"("tenantId", "status");

-- CreateIndex
CREATE INDEX "AuditFinding_type_idx" ON "AuditFinding"("type");

-- CreateIndex
CREATE INDEX "RiskRegister_tenantId_category_idx" ON "RiskRegister"("tenantId", "category");

-- CreateIndex
CREATE INDEX "RiskRegister_tenantId_status_idx" ON "RiskRegister"("tenantId", "status");

-- CreateIndex
CREATE INDEX "RiskRegister_riskScore_idx" ON "RiskRegister"("riskScore");

-- CreateIndex
CREATE INDEX "Contract_tenantId_status_idx" ON "Contract"("tenantId", "status");

-- CreateIndex
CREATE INDEX "Contract_tenantId_type_idx" ON "Contract"("tenantId", "type");

-- CreateIndex
CREATE INDEX "Contract_endDate_idx" ON "Contract"("endDate");

-- CreateIndex
CREATE INDEX "HcmEmployee_tenantId_department_idx" ON "HcmEmployee"("tenantId", "department");

-- CreateIndex
CREATE INDEX "HcmEmployee_tenantId_status_idx" ON "HcmEmployee"("tenantId", "status");

-- CreateIndex
CREATE UNIQUE INDEX "HcmEmployee_tenantId_employeeId_key" ON "HcmEmployee"("tenantId", "employeeId");

-- CreateIndex
CREATE UNIQUE INDEX "HcmEmployee_tenantId_email_key" ON "HcmEmployee"("tenantId", "email");

-- CreateIndex
CREATE INDEX "MarketingCampaign_tenantId_status_idx" ON "MarketingCampaign"("tenantId", "status");

-- CreateIndex
CREATE INDEX "MarketingCampaign_tenantId_type_idx" ON "MarketingCampaign"("tenantId", "type");

-- CreateIndex
CREATE INDEX "MarketingCampaign_startDate_endDate_idx" ON "MarketingCampaign"("startDate", "endDate");

-- CreateIndex
CREATE INDEX "Partner_tenantId_type_idx" ON "Partner"("tenantId", "type");

-- CreateIndex
CREATE INDEX "Partner_tenantId_status_idx" ON "Partner"("tenantId", "status");

-- CreateIndex
CREATE INDEX "Partner_tier_idx" ON "Partner"("tier");

-- CreateIndex
CREATE UNIQUE INDEX "Partner_tenantId_name_key" ON "Partner"("tenantId", "name");

-- CreateIndex
CREATE INDEX "ProcessWorkflow_tenantId_category_idx" ON "ProcessWorkflow"("tenantId", "category");

-- CreateIndex
CREATE INDEX "ProcessWorkflow_status_idx" ON "ProcessWorkflow"("status");

-- CreateIndex
CREATE UNIQUE INDEX "ProcessWorkflow_tenantId_name_version_key" ON "ProcessWorkflow"("tenantId", "name", "version");

-- CreateIndex
CREATE INDEX "QualityCheck_tenantId_type_idx" ON "QualityCheck"("tenantId", "type");

-- CreateIndex
CREATE INDEX "QualityCheck_tenantId_status_idx" ON "QualityCheck"("tenantId", "status");

-- CreateIndex
CREATE INDEX "QualityCheck_standard_idx" ON "QualityCheck"("standard");

-- CreateIndex
CREATE INDEX "KnowledgeArticle_tenantId_category_idx" ON "KnowledgeArticle"("tenantId", "category");

-- CreateIndex
CREATE INDEX "KnowledgeArticle_tenantId_status_idx" ON "KnowledgeArticle"("tenantId", "status");

-- CreateIndex
CREATE INDEX "KnowledgeArticle_tags_idx" ON "KnowledgeArticle" USING GIN ("tags");

-- CreateIndex
CREATE INDEX "VendorRecord_tenantId_category_idx" ON "VendorRecord"("tenantId", "category");

-- CreateIndex
CREATE INDEX "VendorRecord_tenantId_status_idx" ON "VendorRecord"("tenantId", "status");

-- CreateIndex
CREATE INDEX "VendorRecord_riskRating_idx" ON "VendorRecord"("riskRating");

-- CreateIndex
CREATE UNIQUE INDEX "VendorRecord_tenantId_name_key" ON "VendorRecord"("tenantId", "name");

-- CreateIndex
CREATE INDEX "DataAsset_tenantId_type_idx" ON "DataAsset"("tenantId", "type");

-- CreateIndex
CREATE INDEX "DataAsset_tenantId_classification_idx" ON "DataAsset"("tenantId", "classification");

-- CreateIndex
CREATE INDEX "DataAsset_domain_idx" ON "DataAsset"("domain");

-- CreateIndex
CREATE UNIQUE INDEX "DataAsset_tenantId_name_key" ON "DataAsset"("tenantId", "name");

-- CreateIndex
CREATE INDEX "Device_tenantId_type_idx" ON "Device"("tenantId", "type");

-- CreateIndex
CREATE INDEX "Device_tenantId_status_idx" ON "Device"("tenantId", "status");

-- CreateIndex
CREATE INDEX "Device_lastSeen_idx" ON "Device"("lastSeen");

-- CreateIndex
CREATE UNIQUE INDEX "Device_tenantId_serialNumber_key" ON "Device"("tenantId", "serialNumber");

-- CreateIndex
CREATE INDEX "IntegrationConfig_tenantId_type_idx" ON "IntegrationConfig"("tenantId", "type");

-- CreateIndex
CREATE INDEX "IntegrationConfig_tenantId_status_idx" ON "IntegrationConfig"("tenantId", "status");

-- CreateIndex
CREATE INDEX "IntegrationConfig_provider_idx" ON "IntegrationConfig"("provider");

-- CreateIndex
CREATE UNIQUE INDEX "IntegrationConfig_tenantId_name_key" ON "IntegrationConfig"("tenantId", "name");

-- CreateIndex
CREATE INDEX "SupportTicket_tenantId_status_idx" ON "SupportTicket"("tenantId", "status");

-- CreateIndex
CREATE INDEX "SupportTicket_tenantId_priority_idx" ON "SupportTicket"("tenantId", "priority");

-- CreateIndex
CREATE INDEX "SupportTicket_assignee_idx" ON "SupportTicket"("assignee");

-- CreateIndex
CREATE UNIQUE INDEX "SupportTicket_tenantId_ticketNumber_key" ON "SupportTicket"("tenantId", "ticketNumber");

-- CreateIndex
CREATE INDEX "DocumentRecord_tenantId_type_idx" ON "DocumentRecord"("tenantId", "type");

-- CreateIndex
CREATE INDEX "DocumentRecord_tenantId_status_idx" ON "DocumentRecord"("tenantId", "status");

-- CreateIndex
CREATE INDEX "DocumentRecord_tags_idx" ON "DocumentRecord" USING GIN ("tags");

-- CreateIndex
CREATE INDEX "NotificationTemplate_tenantId_channel_idx" ON "NotificationTemplate"("tenantId", "channel");

-- CreateIndex
CREATE INDEX "NotificationTemplate_status_idx" ON "NotificationTemplate"("status");

-- CreateIndex
CREATE UNIQUE INDEX "NotificationTemplate_tenantId_name_channel_locale_key" ON "NotificationTemplate"("tenantId", "name", "channel", "locale");

-- CreateIndex
CREATE INDEX "OpenBankingApi_tenantId_standard_idx" ON "OpenBankingApi"("tenantId", "standard");

-- CreateIndex
CREATE INDEX "OpenBankingApi_status_idx" ON "OpenBankingApi"("status");

-- CreateIndex
CREATE UNIQUE INDEX "OpenBankingApi_tenantId_name_version_key" ON "OpenBankingApi"("tenantId", "name", "version");

-- CreateIndex
CREATE INDEX "OpenFinanceProduct_tenantId_category_idx" ON "OpenFinanceProduct"("tenantId", "category");

-- CreateIndex
CREATE INDEX "OpenFinanceProduct_tenantId_provider_idx" ON "OpenFinanceProduct"("tenantId", "provider");

-- CreateIndex
CREATE INDEX "OpenInsurancePolicy_tenantId_type_idx" ON "OpenInsurancePolicy"("tenantId", "type");

-- CreateIndex
CREATE INDEX "OpenInsurancePolicy_tenantId_status_idx" ON "OpenInsurancePolicy"("tenantId", "status");

-- CreateIndex
CREATE UNIQUE INDEX "OpenInsurancePolicy_tenantId_policyNumber_key" ON "OpenInsurancePolicy"("tenantId", "policyNumber");

-- CreateIndex
CREATE INDEX "OpenDataDataset_tenantId_category_idx" ON "OpenDataDataset"("tenantId", "category");

-- CreateIndex
CREATE INDEX "OpenDataDataset_format_idx" ON "OpenDataDataset"("format");

-- CreateIndex
CREATE INDEX "OpenHealthResource_tenantId_resourceType_idx" ON "OpenHealthResource"("tenantId", "resourceType");

-- CreateIndex
CREATE INDEX "OpenHealthResource_fhirVersion_idx" ON "OpenHealthResource"("fhirVersion");

-- CreateIndex
CREATE INDEX "OpenHealthResource_subject_idx" ON "OpenHealthResource"("subject");

-- CreateIndex
CREATE INDEX "OpenEducationCourse_tenantId_format_idx" ON "OpenEducationCourse"("tenantId", "format");

-- CreateIndex
CREATE INDEX "OpenEducationCourse_tenantId_provider_idx" ON "OpenEducationCourse"("tenantId", "provider");

-- CreateIndex
CREATE INDEX "OpenEducationCourse_level_idx" ON "OpenEducationCourse"("level");

-- CreateIndex
CREATE INDEX "OpenInnovationChallenge_tenantId_type_idx" ON "OpenInnovationChallenge"("tenantId", "type");

-- CreateIndex
CREATE INDEX "OpenInnovationChallenge_tenantId_status_idx" ON "OpenInnovationChallenge"("tenantId", "status");

-- CreateIndex
CREATE INDEX "OpenInnovationChallenge_endDate_idx" ON "OpenInnovationChallenge"("endDate");

-- CreateIndex
CREATE INDEX "InnovationProject_tenantId_type_idx" ON "InnovationProject"("tenantId", "type");

-- CreateIndex
CREATE INDEX "InnovationProject_tenantId_status_idx" ON "InnovationProject"("tenantId", "status");

-- CreateIndex
CREATE INDEX "InnovationProject_stage_idx" ON "InnovationProject"("stage");

-- CreateIndex
CREATE UNIQUE INDEX "InnovationProject_tenantId_name_key" ON "InnovationProject"("tenantId", "name");

-- AddForeignKey
ALTER TABLE "capabilities" ADD CONSTRAINT "capabilities_parent_id_fkey" FOREIGN KEY ("parent_id") REFERENCES "capabilities"("id") ON DELETE CASCADE ON UPDATE NO ACTION;

-- AddForeignKey
ALTER TABLE "capability_geography" ADD CONSTRAINT "capability_geography_capability_id_fkey" FOREIGN KEY ("capability_id") REFERENCES "capabilities"("id") ON DELETE CASCADE ON UPDATE NO ACTION;

-- AddForeignKey
ALTER TABLE "capability_geography" ADD CONSTRAINT "capability_geography_country_code_fkey" FOREIGN KEY ("country_code") REFERENCES "countries"("code") ON DELETE CASCADE ON UPDATE NO ACTION;

-- AddForeignKey
ALTER TABLE "capability_industries" ADD CONSTRAINT "capability_industries_capability_id_fkey" FOREIGN KEY ("capability_id") REFERENCES "capabilities"("id") ON DELETE CASCADE ON UPDATE NO ACTION;

-- AddForeignKey
ALTER TABLE "capability_industries" ADD CONSTRAINT "capability_industries_industry_id_fkey" FOREIGN KEY ("industry_id") REFERENCES "industries"("id") ON DELETE CASCADE ON UPDATE NO ACTION;

-- AddForeignKey
ALTER TABLE "capability_modules" ADD CONSTRAINT "capability_modules_capability_id_fkey" FOREIGN KEY ("capability_id") REFERENCES "capabilities"("id") ON DELETE CASCADE ON UPDATE NO ACTION;

-- AddForeignKey
ALTER TABLE "capability_modules" ADD CONSTRAINT "capability_modules_module_id_fkey" FOREIGN KEY ("module_id") REFERENCES "modules"("id") ON DELETE CASCADE ON UPDATE NO ACTION;

-- AddForeignKey
ALTER TABLE "capability_organizational" ADD CONSTRAINT "capability_organizational_capability_id_fkey" FOREIGN KEY ("capability_id") REFERENCES "capabilities"("id") ON DELETE CASCADE ON UPDATE NO ACTION;

-- AddForeignKey
ALTER TABLE "capability_performance" ADD CONSTRAINT "capability_performance_capability_id_fkey" FOREIGN KEY ("capability_id") REFERENCES "capabilities"("id") ON DELETE CASCADE ON UPDATE NO ACTION;

-- AddForeignKey
ALTER TABLE "capability_strategy" ADD CONSTRAINT "capability_strategy_capability_id_fkey" FOREIGN KEY ("capability_id") REFERENCES "capabilities"("id") ON DELETE CASCADE ON UPDATE NO ACTION;

-- AddForeignKey
ALTER TABLE "ChatMessage" ADD CONSTRAINT "ChatMessage_sessionId_fkey" FOREIGN KEY ("sessionId") REFERENCES "ChatSession"("id") ON DELETE CASCADE ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "HealthcareAppointment" ADD CONSTRAINT "HealthcareAppointment_patientId_fkey" FOREIGN KEY ("patientId") REFERENCES "HealthcarePatient"("id") ON DELETE RESTRICT ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "FinancialTransaction" ADD CONSTRAINT "FinancialTransaction_accountId_fkey" FOREIGN KEY ("accountId") REFERENCES "FinancialAccount"("id") ON DELETE RESTRICT ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "GovernmentServiceRequest" ADD CONSTRAINT "GovernmentServiceRequest_citizenId_fkey" FOREIGN KEY ("citizenId") REFERENCES "GovernmentCitizen"("id") ON DELETE RESTRICT ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "EducationEnrollment" ADD CONSTRAINT "EducationEnrollment_studentId_fkey" FOREIGN KEY ("studentId") REFERENCES "EducationStudent"("id") ON DELETE RESTRICT ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "EducationEnrollment" ADD CONSTRAINT "EducationEnrollment_courseId_fkey" FOREIGN KEY ("courseId") REFERENCES "EducationCourse"("id") ON DELETE RESTRICT ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "RetailOrderItem" ADD CONSTRAINT "RetailOrderItem_orderId_fkey" FOREIGN KEY ("orderId") REFERENCES "RetailOrder"("id") ON DELETE RESTRICT ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "RetailOrderItem" ADD CONSTRAINT "RetailOrderItem_productId_fkey" FOREIGN KEY ("productId") REFERENCES "RetailProduct"("id") ON DELETE RESTRICT ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "IamUserRole" ADD CONSTRAINT "IamUserRole_userId_fkey" FOREIGN KEY ("userId") REFERENCES "IamUser"("id") ON DELETE CASCADE ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "IamUserRole" ADD CONSTRAINT "IamUserRole_roleId_fkey" FOREIGN KEY ("roleId") REFERENCES "IamRole"("id") ON DELETE CASCADE ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "IamSession" ADD CONSTRAINT "IamSession_userId_fkey" FOREIGN KEY ("userId") REFERENCES "IamUser"("id") ON DELETE CASCADE ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "IamMfaDevice" ADD CONSTRAINT "IamMfaDevice_userId_fkey" FOREIGN KEY ("userId") REFERENCES "IamUser"("id") ON DELETE CASCADE ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "IamApiKey" ADD CONSTRAINT "IamApiKey_userId_fkey" FOREIGN KEY ("userId") REFERENCES "IamUser"("id") ON DELETE SET NULL ON UPDATE CASCADE;

