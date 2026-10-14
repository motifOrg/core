INSERT INTO campaigns (
    id,
    name,
    description,
    status,
    workflow,
    gmail_label,
    created_at,
    updated_at
) VALUES (
    '8be4df61-93ca-11d2-aa0d-00e098032b8c',
    'Bootstrapped B2B SaaS',
    'Founder-owned, capital-efficient software companies with self-serve subscription revenue in North America and Europe.',
    'active',
    'both',
    'B2B SaaS',
    NOW(),
    NOW()
)
ON CONFLICT (id) DO NOTHING;

INSERT INTO icp_criteria (
    campaign_id,
    thesis,
    industries,
    geographies,
    business_models,
    characteristics,
    exclusions,
    revenue_min,
    revenue_max,
    ebitda_min,
    employees_min,
    employees_max,
    ownership
) VALUES (
    '8be4df61-93ca-11d2-aa0d-00e098032b8c',
    'Acquire a profitable, founder-run B2B SaaS business that grew without venture capital, keep the product and team, and add sales and partnerships the founders never built.',
    '{"B2B SaaS","Productivity software","Marketing software","Analytics software","Developer tools"}',
    '{"United States","Canada","United Kingdom","European Union"}',
    '{"Subscription SaaS","Usage-based SaaS"}',
    '{"Bootstrapped or minimal outside capital","Founder still runs the business","Self-serve sales motion","Profitable or near break-even","Small remote team"}',
    '{"Venture-backed beyond seed","Already acquired or PE-owned","Consumer apps","Hardware"}',
    2000000,
    40000000,
    500000,
    5,
    150,
    'Founder owned, bootstrapped'
)
ON CONFLICT (campaign_id) DO NOTHING;