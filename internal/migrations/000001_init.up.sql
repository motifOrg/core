CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE campaigns (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL DEFAULT 'draft'
        CHECK (status IN ('draft', 'active', 'paused', 'archived')),
    workflow TEXT NOT NULL DEFAULT 'outbound'
        CHECK (workflow IN ('outbound', 'inbound', 'both')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE icp_criteria (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    campaign_id UUID NOT NULL UNIQUE REFERENCES campaigns(id) ON DELETE CASCADE,
    thesis TEXT NOT NULL DEFAULT '',
    industries TEXT[] NOT NULL DEFAULT '{}',
    geographies TEXT[] NOT NULL DEFAULT '{}',
    business_models TEXT[] NOT NULL DEFAULT '{}',
    characteristics TEXT[] NOT NULL DEFAULT '{}',
    exclusions TEXT[] NOT NULL DEFAULT '{}',
    revenue_min BIGINT CHECK (revenue_min IS NULL OR revenue_min >= 0),
    revenue_max BIGINT CHECK (revenue_max IS NULL OR revenue_max >= 0),
    ebitda_min BIGINT CHECK (ebitda_min IS NULL OR ebitda_min >= 0),
    employees_min INT CHECK (employees_min IS NULL OR employees_min >= 0),
    employees_max INT CHECK (employees_max IS NULL OR employees_max >= 0),
    ownership TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CHECK (revenue_min IS NULL OR revenue_max IS NULL OR revenue_min <= revenue_max),
    CHECK (employees_min IS NULL OR employees_max IS NULL OR employees_min <= employees_max)
);

CREATE TABLE documents (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    campaign_id UUID NOT NULL REFERENCES campaigns(id) ON DELETE CASCADE,
    filename TEXT NOT NULL,
    storage_key TEXT NOT NULL,
    mime_type TEXT NOT NULL,
    size_bytes BIGINT NOT NULL DEFAULT 0,
    status TEXT NOT NULL DEFAULT 'uploaded'
        CHECK (status IN ('uploaded', 'processing', 'processed', 'failed')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE companies (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    campaign_id UUID NOT NULL REFERENCES campaigns(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    domain TEXT NOT NULL DEFAULT '',
    industry TEXT NOT NULL DEFAULT '',
    location TEXT NOT NULL DEFAULT '',
    description TEXT NOT NULL DEFAULT '',
    employees INT,
    revenue_estimate TEXT NOT NULL DEFAULT '',
    founded_year INT,
    owner_name TEXT NOT NULL DEFAULT '',
    source TEXT NOT NULL DEFAULT 'manual'
        CHECK (source IN ('sample', 'csv', 'manual', 'inbound')),
    is_sample BOOLEAN NOT NULL DEFAULT false,
    fit_score INT CHECK (fit_score IS NULL OR fit_score BETWEEN 0 AND 100),
    fit_summary TEXT NOT NULL DEFAULT '',
    priority TEXT NOT NULL DEFAULT 'medium'
        CHECK (priority IN ('high', 'medium', 'low')),
    research_status TEXT NOT NULL DEFAULT 'not_started'
        CHECK (research_status IN ('not_started', 'in_progress', 'complete')),
    evaluation_status TEXT NOT NULL DEFAULT 'new'
        CHECK (evaluation_status IN ('new', 'shortlisted', 'contacted', 'in_discussion', 'rejected')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_companies_campaign_domain
    ON companies(campaign_id, lower(domain))
    WHERE domain <> '';

CREATE TABLE research_findings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id UUID NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    kind TEXT NOT NULL
        CHECK (kind IN ('fit', 'finding', 'risk', 'missing', 'next_action')),
    title TEXT NOT NULL,
    content TEXT NOT NULL DEFAULT '',
    source_url TEXT NOT NULL DEFAULT '',
    confidence TEXT NOT NULL DEFAULT 'medium'
        CHECK (confidence IN ('high', 'medium', 'low')),
    evidence_status TEXT NOT NULL DEFAULT 'unverified'
        CHECK (evidence_status IN ('sample', 'unverified', 'verified')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE research_messages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id UUID NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    body TEXT NOT NULL,
    sender_role TEXT NOT NULL CHECK (sender_role IN ('ai', 'user', 'system')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE inbound_opportunities (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    campaign_id UUID NOT NULL REFERENCES campaigns(id) ON DELETE CASCADE,
    company_id UUID REFERENCES companies(id) ON DELETE SET NULL,
    source TEXT NOT NULL DEFAULT 'manual'
        CHECK (source IN ('gmail', 'manual', 'sample')),
    external_ref TEXT NOT NULL DEFAULT '',
    sender_name TEXT NOT NULL DEFAULT '',
    sender_email TEXT NOT NULL,
    subject TEXT NOT NULL DEFAULT '',
    body TEXT NOT NULL DEFAULT '',
    company_name TEXT NOT NULL DEFAULT '',
    company_domain TEXT NOT NULL DEFAULT '',
    received_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    is_sample BOOLEAN NOT NULL DEFAULT false,
    fit_score INT CHECK (fit_score IS NULL OR fit_score BETWEEN 0 AND 100),
    fit_summary TEXT NOT NULL DEFAULT '',
    priority TEXT NOT NULL DEFAULT 'medium'
        CHECK (priority IN ('high', 'medium', 'low')),
    status TEXT NOT NULL DEFAULT 'new'
        CHECK (status IN ('new', 'reviewing', 'qualified', 'declined')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_inbound_external_ref
    ON inbound_opportunities(source, external_ref)
    WHERE external_ref <> '';

CREATE TABLE message_drafts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id UUID REFERENCES companies(id) ON DELETE CASCADE,
    opportunity_id UUID REFERENCES inbound_opportunities(id) ON DELETE CASCADE,
    recipient TEXT NOT NULL DEFAULT '',
    subject TEXT NOT NULL DEFAULT '',
    body TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL DEFAULT 'draft'
        CHECK (status IN ('draft', 'reviewed')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CHECK ((company_id IS NULL) <> (opportunity_id IS NULL))
);

CREATE INDEX idx_companies_campaign ON companies(campaign_id);
CREATE INDEX idx_companies_fit ON companies(campaign_id, fit_score DESC NULLS LAST);
CREATE INDEX idx_documents_campaign ON documents(campaign_id);
CREATE INDEX idx_findings_company ON research_findings(company_id);
CREATE INDEX idx_messages_company ON research_messages(company_id, created_at);
CREATE INDEX idx_inbound_campaign ON inbound_opportunities(campaign_id, received_at DESC);
CREATE INDEX idx_drafts_company ON message_drafts(company_id);
CREATE INDEX idx_drafts_opportunity ON message_drafts(opportunity_id);
