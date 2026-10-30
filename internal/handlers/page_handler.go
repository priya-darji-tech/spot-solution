package handlers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"spot-solution/internal/config"
)

type ServiceDetail struct {
	Slug         string
	Title        string
	ShortDesc    string
	Icon         string
	HeroHeading  string
	FullDesc     string
	Features     []string
	Benefits     []string
	Deliverables []string
}

var ServicesCatalog = map[string]ServiceDetail{
	"custom-software": {
		Slug:        "custom-software",
		Title:       "Custom Software Development",
		ShortDesc:   "Creating robust enterprise solutions for any business challenge or opportunity. Full solution, from envisioning and requirements through development to ongoing support.",
		Icon:        "fas fa-code",
		HeroHeading: "Custom Software Solutions Engineered for Your Business Operations",
		FullDesc:    "We’ve been in business over 20 years and our reputation has kept us around. Spot Solutions creates Custom Software, Web Application Development and Legacy Upgrades that are solutions to your business problems. We take great pride in tracking our metrics and building software that companies both actually need and work as intended.",
		Features: []string{
			"Modular, clean-architecture software tailored to proprietary business models",
			"High-concurrency backend services with Go, .NET, PostgreSQL, and modern APIs",
			"Strict domain modeling and automated test-driven test coverage",
			"Legacy data migration and automated modernization pipelines",
		},
		Benefits: []string{
			"Eliminate expensive recurring third-party licensing fees",
			"Attain 100% control over features, intellectual property, and user journeys",
			"Seamless integration with your existing databases, ERPs, and cloud services",
		},
		Deliverables: []string{
			"Complete Source Code & Documentation",
			"Automated CI/CD Pipelines & Docker Containers",
			"Comprehensive User Manuals & SLA Support",
		},
	},
	"sharepoint": {
		Slug:        "sharepoint",
		Title:       "SharePoint & Microsoft 365",
		ShortDesc:   "Experts at SharePoint customization, collaboration portals, migrations, and extensive custom development.",
		Icon:        "fas fa-share-alt",
		HeroHeading: "Microsoft Solutions Partner for SharePoint Online and Hybrid Environments",
		FullDesc:    "From design and implementation to user adoption and training, our experienced M365 and SharePoint consultants and developers deliver the right solution to achieve your goals. We build intranets people actually use.",
		Features: []string{
			"SharePoint Online and hybrid intranet deployment",
			"SPFx (SharePoint Framework) custom web parts and extensions",
			"ShareGate content and security migration tooling",
			"Document management taxonomy, metadata, and governance",
		},
		Benefits: []string{
			"Accelerate corporate communication and document discoverability",
			"Ensure compliance and robust access controls",
			"Seamless integration with Microsoft Teams and Office apps",
		},
		Deliverables: []string{
			"Configured SharePoint Intranet Portal",
			"Information Architecture & Governance Plan",
			"Staff Adoption & Admin Training Sessions",
		},
	},
	"sharepoint-migration": {
		Slug:        "sharepoint-migration",
		Title:       "SharePoint Migration",
		ShortDesc:   "Migration projects can be overwhelming. We’ve moved 49 organizations — including the BC Ministry of Education, Washington Companies and Teck — to SharePoint Online with zero lost content.",
		Icon:        "fas fa-exchange-alt",
		HeroHeading: "SharePoint Migration Services in Vancouver & Calgary",
		FullDesc:    "Our SharePoint migration experts work with your team to understand and design a migration solution for SharePoint on-premises, online, and hybrid. Typical timeline: 6–14 weeks. Whether migrating from a prior SharePoint version, .NET, ASP, Lotus Notes, or another system, we help every step of the way.",
		Features: []string{
			"SharePoint version upgrades to modern Online / Subscription Edition",
			"Microsoft 365 migration from on-premises or hybrid environments",
			"Data centre modernization and cloud relocation",
			"Legacy system migration and multi-farm consolidation",
			"Deployment planning for on-premises, cloud, or hybrid",
		},
		Benefits: []string{
			"Zero lost content with proven enterprise migration track record",
			"Seamless cutover with change management and training included",
			"Reduced farm maintenance costs through consolidation",
		},
		Deliverables: []string{
			"Analysis & Planning Report",
			"Migration Design & Governance Model",
			"Validated Cutover with UAT Sign-off",
		},
	},
	"sharepoint-migration-2": {
		Slug:        "sharepoint-migration",
		Title:       "SharePoint Migration",
		ShortDesc:   "Migration projects can be overwhelming. We’ve moved 49 organizations — including the BC Ministry of Education, Washington Companies and Teck — to SharePoint Online with zero lost content.",
		Icon:        "fas fa-exchange-alt",
		HeroHeading: "SharePoint Migration Services in Vancouver & Calgary",
		FullDesc:    "Our SharePoint migration experts work with your team to understand and design a migration solution for SharePoint on-premises, online, and hybrid. Typical timeline: 6–14 weeks.",
		Features: []string{
			"SharePoint version upgrades to modern Online / Subscription Edition",
			"Microsoft 365 migration from on-premises or hybrid environments",
			"Data centre modernization and cloud relocation",
			"Legacy system migration and multi-farm consolidation",
		},
		Benefits: []string{
			"Zero lost content with proven enterprise migration track record",
			"Seamless cutover with change management and training included",
			"Reduced farm maintenance costs through consolidation",
		},
		Deliverables: []string{
			"Analysis & Planning Report",
			"Migration Design & Governance Model",
			"Validated Cutover with UAT Sign-off",
		},
	},
	"miscrosoft-365-experts": {
		Slug:        "miscrosoft-365-experts",
		Title:       "Microsoft 365 Experts",
		ShortDesc:   "Maximize productivity and collaboration with expert Microsoft 365 planning, deployment, Copilot enablement, and ongoing support.",
		Icon:        "fas fa-users",
		HeroHeading: "Work Better Together with Microsoft 365",
		FullDesc:    "Productivity and collaboration is a mandatory ingredient for success. Spot Solutions is an experienced cloud service provider that helps you architect the best Microsoft 365 plans and modules to achieve your targets. We identify, configure, deploy, and manage your Microsoft 365 platforms to maximize your investment.",
		Features: []string{
			"Microsoft 365 Copilot enablement across Word, Excel, SharePoint, Outlook, and Teams",
			"Group chat, online meetings, calling, and web conferencing",
			"Unified email, calendar, contacts, and tasks",
			"Custom automations, reporting tools, and Copilot agents",
			"Secure authentication and automated policy-based access rules",
		},
		Benefits: []string{
			"Free 24×7×365 partner support access",
			"Free migration and onboarding assistance",
			"No long-term commitment — competitive Microsoft rates",
			"Trusted methodologies from Microsoft 365 experts",
		},
		Deliverables: []string{
			"M365 Tenant Configuration & Security Audit",
			"Copilot Readiness & Governance Plan",
			"Migration & Onboarding Support",
		},
	},
	"powerplatform": {
		Slug:        "powerplatform",
		Title:       "Power Platform",
		ShortDesc:   "Build low-code apps, automate approvals, and turn data into dashboards with Power Apps, Power Automate, Power BI, and Copilot Studio.",
		Icon:        "fas fa-bolt",
		HeroHeading: "Power Platform Consulting in Vancouver & Calgary",
		FullDesc:    "We help organizations across BC and Alberta build low-code apps, automate approvals, and turn their data into dashboards with Power Apps, Power Automate, Power BI, and Copilot Studio. It’s part of our SharePoint & Microsoft 365 consulting, so your apps connect cleanly to the systems your team already uses.",
		Features: []string{
			"Power Apps development connected to SharePoint, SQL Server, or Dataverse",
			"Power BI analytics and live executive dashboards",
			"Power Automate flows for approvals, reminders, and file routing",
			"Power Platform integration with Microsoft 365, Dynamics 365, Azure, and SQL",
			"Copilot Studio agents grounded in your SharePoint content",
		},
		Benefits: []string{
			"Customized solutions tailored to your workflows",
			"Streamlined automation that cuts manual effort",
			"Seamless integration with existing Microsoft investments",
			"Microsoft Power Platform certified consultants",
		},
		Deliverables: []string{
			"Deployed Power Apps Application",
			"Automated Workflows & Connectors",
			"Dashboards, Governance & Admin Handoff",
		},
	},
	"intranet": {
		Slug:        "intranet",
		Title:       "SharePoint Intranet",
		ShortDesc:   "Custom SharePoint Intranet for companies that don’t fit in a box — built around your processes, branding, and people.",
		Icon:        "fas fa-network-wired",
		HeroHeading: "Build YOUR Intranet to Suit YOUR Business",
		FullDesc:    "Here at Spot, we want to understand your business processes to build an intranet solution that will fit your needs. Our team of designers, business analysts, and developers make an intuitive and mobile Intranet that speaks to your needs — not an in-a-box product that tells you how to run your business.",
		Features: []string{
			"Document management with libraries, permissions, and departmental spaces",
			"Collaboration with directories, co-editing, and team workspaces",
			"Communication via comments, forums, message boards, and social feeds",
			"Business automation with approval workflows and information routing",
			"Mobile compatibility and bespoke branded design",
		},
		Benefits: []string{
			"Intranet that matches how you actually work",
			"Faster document findability and fewer email trails",
			"Digitized processes with clear visibility and ownership",
		},
		Deliverables: []string{
			"Current State Assessment",
			"Future State Planning & Governance",
			"Configured Intranet with Training & Closure",
		},
	},
	"sharepoint-custom-development": {
		Slug:        "sharepoint-custom-development",
		Title:       "SharePoint Custom Development",
		ShortDesc:   "We build SharePoint systems that do what you need them to do — beyond out-of-the-box configuration limits.",
		Icon:        "fas fa-puzzle-piece",
		HeroHeading: "SharePoint: Custom Development",
		FullDesc:    "SharePoint is incredibly flexible, but sometimes simple configuration is not enough. Spot Solutions creates custom solutions for unique business needs — intranet, extranet, and internet-based systems that extend SharePoint beyond built-in limits.",
		Features: []string{
			"SharePoint content and enterprise document management systems",
			"Multi-lingual content management",
			"SharePoint farm architecture per Microsoft best practices",
			"SharePoint app development (provider-hosted and SharePoint-hosted)",
			"Azure AD sync, LOB integrations, BI dashboards, and workflows",
			"ShareGate / Metalogix / MetaVis migrations to M365",
		},
		Benefits: []string{
			"You set the limits — unique experiences and tools for your business",
			"Extend SharePoint beyond configuration-only constraints",
			"Integrate line-of-business apps with M365 solutions",
		},
		Deliverables: []string{
			"Custom SharePoint / SPFx Applications",
			"Integration & Migration Packages",
			"Deployment Documentation & Support",
		},
	},
	"sharepoint-support": {
		Slug:        "sharepoint-support",
		Title:       "SharePoint Support",
		ShortDesc:   "Proactive SharePoint support packages with helpdesk access, SLA response times, and flexible hours valid for 12 months.",
		Icon:        "fas fa-headset",
		HeroHeading: "Get the SharePoint Support You Need",
		FullDesc:    "Spot Solutions clearly understands the challenges of supporting a SharePoint environment. We create a customized procedure so your applications stay tuned and meet key business challenges. Support hours are valid for 12 months and can be used for many services — not just break/fix. Our support desk is onshore and we can come on-site.",
		Features: []string{
			"Proactive management with helpdesk access and SLA response times",
			"Flexible hours usable across products and services",
			"Scale resources up or down without hiring full-time experts",
			"Pre-built solutions including training and change management",
			"Agile delivery focused on your business requirements",
		},
		Benefits: []string{
			"More control over support spending",
			"Onshore Canadian support desk with on-site options",
			"Access to experts when you need them",
		},
		Deliverables: []string{
			"Support Agreement & Service Level Guide",
			"Helpdesk Access & Response Procedures",
			"Quarterly Health Checks (as agreed)",
		},
	},
	"ai": {
		Slug:        "ai",
		Title:       "Artificial Intelligence (AI) & Machine Learning",
		ShortDesc:   "Incorporate practical AI tools, Copilot Studio agents, and automated data processing into your business.",
		Icon:        "fas fa-robot",
		HeroHeading: "Practical Enterprise AI: Move Beyond Hype to Measurable ROI",
		FullDesc:    "As AI tools continue to develop, Spot Solutions guides you on incorporating them into your business processes, strategies, and workflows. We build custom agents and chatbots to improve efficiency.",
		Features: []string{
			"Custom Copilot Studio agents for internal knowledge retrieval",
			"Retrieval-Augmented Generation (RAG) over proprietary enterprise documents",
			"Automated document classification and OCR extraction",
			"Enterprise data privacy guarantees: zero model training on your data",
		},
		Benefits: []string{
			"Save hundreds of staff hours spent searching through file repositories",
			"Provide instant 24/7 intelligent answers to employee queries",
			"Accelerate decision-making with surfaced insights",
		},
		Deliverables: []string{
			"Calibrated AI Agent & Connectors",
			"Security & Permissions Scoping Report",
			"Performance & Accuracy Evaluation Metrics",
		},
	},
	"mobile": {
		Slug:        "mobile",
		Title:       "Mobile Solutions",
		ShortDesc:   "Native and cross-platform enterprise mobile applications focused on field productivity and security.",
		Icon:        "fas fa-mobile-alt",
		HeroHeading: "Enterprise Mobile Applications Engineered for Frontline Efficiency",
		FullDesc:    "Building mobile solutions with a focus on enterprise productivity. Native and hybrid mobile application development for Android and iOS platforms with offline sync capabilities.",
		Features: []string{
			"Cross-platform Flutter / React Native and native mobile engineering",
			"Offline-first local database caching and automatic cloud sync",
			"Biometric authentication and secure device hardware integration",
			"Custom push notifications and real-time alerts",
		},
		Benefits: []string{
			"Empower remote and field personnel with immediate access to data",
			"Capture signatures, photos, and inspection notes on-site",
			"Improve customer response times and operational accuracy",
		},
		Deliverables: []string{
			"Compiled iOS and Android App Packages",
			"Mobile Backend APIs & Database Connectors",
			"App Store & MDM Deployment Documentation",
		},
	},
	"web-applications": {
		Slug:        "web-applications",
		Title:       "Web Applications",
		ShortDesc:   "We build Web Applications tailored specifically for your business needs, from intranet tools to public portals.",
		Icon:        "fas fa-globe",
		HeroHeading: "High-Performance Web Applications Built for Reliability and Scale",
		FullDesc:    "From staff-only or client-only intranet applications to public-facing full-functionality Internet web applications, we engineer solutions with speed, intuitive UX, and bulletproof security.",
		Features: []string{
			"Server-rendered HTML with progressive enhancement for optimal SEO and performance",
			"Accessible interfaces adhering strictly to modern standards",
			"Interactive dashboards with real-time data and role-based permissions",
			"Enterprise identity integrations (SSO, OAuth2, Microsoft Entra ID)",
		},
		Benefits: []string{
			"Sub-second load times across mobile, tablet, and desktop devices",
			"Zero vendor lock-in with open, standards-based architectures",
			"Effortless cross-browser compatibility and device responsiveness",
		},
		Deliverables: []string{
			"Production Web Application Bundle",
			"Responsive Design Assets & Component System",
			"Security Audit & Load Test Reports",
		},
	},
	"requirements": {
		Slug:        "requirements",
		Title:       "Requirements Analysis",
		ShortDesc:   "Documented, actionable, measurable, and testable requirements defined to guarantee project success.",
		Icon:        "fas fa-search",
		HeroHeading: "Define Your Project Scope Right to Guarantee Delivery On-Time and On-Budget",
		FullDesc:    "We offer high-level analysis to make sure your requirements are documented, actionable, measurable, testable, and related to identified business needs before writing a single line of code.",
		Features: []string{
			"Stakeholder interviews and business process mapping",
			"Functional specification documents and screen wireframes",
			"Technical feasibility analysis and architectural blueprints",
			"Risk assessment matrix and phased delivery roadmaps",
		},
		Benefits: []string{
			"Eliminate expensive mid-project scope creep and rework",
			"Align management and technical teams on clear deliverables",
			"Accurately estimate timelines, budgets, and milestones",
		},
		Deliverables: []string{
			"Comprehensive Requirements Specification",
			"Clickable Prototype Wireframes",
			"Fixed-Price Statement of Work (SOW)",
		},
	},
	"change-managment": {
		Slug:        "change-managment",
		Title:       "Change Management",
		ShortDesc:   "Ensure your software investment succeeds with user adoption strategies, training, and stakeholder buy-in.",
		Icon:        "fas fa-sync-alt",
		HeroHeading: "Drive 100% User Adoption for Every Software Implementation",
		FullDesc:    "No matter how well an application is built, it has no value if it is not used. Spot Solutions provides structured change management, communication plans, and comprehensive user training.",
		Features: []string{
			"User impact analysis and readiness assessments",
			"Tailored training curricula for end users and administrators",
			"Change champion networks and internal communication campaigns",
			"Adoption metrics tracking and post-launch feedback loops",
		},
		Benefits: []string{
			"Overcome employee resistance to new software tools",
			"Shorten learning curves and maintain high productivity",
			"Maximize return on technology investment from Day 1",
		},
		Deliverables: []string{
			"Change Management & Communications Plan",
			"Role-Based Training Videos & Quick Reference Cards",
			"Adoption Health Check Report",
		},
	},
	"change-management": {
		Slug:        "change-managment",
		Title:       "Change Management",
		ShortDesc:   "Ensure your software investment succeeds with user adoption strategies, training, and stakeholder buy-in.",
		Icon:        "fas fa-sync-alt",
		HeroHeading: "Drive 100% User Adoption for Every Software Implementation",
		FullDesc:    "No matter how well an application is built, it has no value if it is not used. Spot Solutions provides structured change management, communication plans, and comprehensive user training.",
		Features: []string{
			"User impact analysis and readiness assessments",
			"Tailored training curricula for end users and administrators",
			"Change champion networks and internal communication campaigns",
			"Adoption metrics tracking and post-launch feedback loops",
		},
		Benefits: []string{
			"Overcome employee resistance to new software tools",
			"Shorten learning curves and maintain high productivity",
			"Maximize return on technology investment from Day 1",
		},
		Deliverables: []string{
			"Change Management & Communications Plan",
			"Role-Based Training Videos & Quick Reference Cards",
			"Adoption Health Check Report",
		},
	},
	"database": {
		Slug:        "database",
		Title:       "Databases",
		ShortDesc:   "Organize, secure, and optimize your business information with modern database architectures.",
		Icon:        "fas fa-database",
		HeroHeading: "High-Availability Database Engineering, Optimization, and Migration",
		FullDesc:    "Organize and control your information using a database optimized for your business needs. Migrating between database services or upgrading SQL Server / PostgreSQL, we have the expertise to advance your company.",
		Features: []string{
			"Relational database design (PostgreSQL, MS SQL Server, MySQL)",
			"Performance indexing, query optimization, and dead-lock remediation",
			"Automated failover, point-in-time recovery, and multi-region replication",
			"Data warehouse structuring for real-time reporting",
		},
		Benefits: []string{
			"Eliminate application sluggishness and database timeouts",
			"Guarantee 99.99% data availability and disaster recoverability",
			"Ensure compliance with Canadian privacy and residency standards",
		},
		Deliverables: []string{
			"Optimized Database Schema & Migration Scripts",
			"Performance Benchmark & Indexing Report",
			"Automated Disaster Recovery Runbooks",
		},
	},
	"hosting": {
		Slug:        "hosting",
		Title:       "Enterprise Hosting",
		ShortDesc:   "Dedicated virtual machines, cloud infrastructure, and managed hosting for SharePoint, web apps, and databases.",
		Icon:        "far fa-folder-open",
		HeroHeading: "Secure, Scalable Enterprise Hosting Backed by 24/7 Monitoring",
		FullDesc:    "As part of a complete solution, we offer full enterprise hosting services, including dedicated virtual machines, co-location, and multi-tenancy shared services for SharePoint, web applications, and databases.",
		Features: []string{
			"Canadian data sovereignty compliant infrastructure",
			"Automated SSL/TLS provisioning and hardware firewall protection",
			"24/7 uptime monitoring with proactive incident resolution",
			"Daily automated snapshots and geographically redundant backups",
		},
		Benefits: []string{
			"Reliable performance with guaranteed 99.9% uptime SLA",
			"Freedom from managing server hardware and operating system patches",
			"Dedicated Canadian technical support team",
		},
		Deliverables: []string{
			"Configured Hosting Environment & Access Credentials",
			"SLA Uptime & Support Agreement",
			"Monthly Performance & Security Reports",
		},
	},
	"legacy-system-upgrades": {
		Slug:        "legacy-system-upgrades",
		Title:       "Legacy System Upgrades",
		ShortDesc:   "Modernize outdated legacy code, ColdFusion, and older .NET systems into secure, scalable modern platforms.",
		Icon:        "fas fa-paper-plane",
		HeroHeading: "Modernize Legacy Applications Without Losing Critical Business Logic",
		FullDesc:    "Upgrading your legacy software becomes more necessary as time progresses and technology changes. Let us help you find the optimal time to engage on such a legacy software upgrade project, and assist you with all the details.",
		Features: []string{
			"Legacy codebase decompilation, reverse engineering, and documentation",
			"ColdFusion to modern Go / .NET / PHP migrations",
			"Zero data loss automated ETL database transformation",
			"Phased cutover strategies to prevent business downtime",
		},
		Benefits: []string{
			"Eliminate reliance on retiring developers and obsolete tech stacks",
			"Significantly reduce server licensing and maintenance costs",
			"Unlock modern web interfaces, mobile responsiveness, and API integrations",
		},
		Deliverables: []string{
			"Legacy System Technical Audit & Migration Roadmap",
			"Modernized Application Codebase & Tests",
			"Zero-Downtime Data Migration Plan",
		},
	},
	"cloud-solutions": {
		Slug:        "cloud-solutions",
		Title:       "Cloud Solutions & Infrastructure",
		ShortDesc:   "Resilient cloud architecture, automated migration, cost governance, and 24/7 reliability on modern infrastructure.",
		Icon:        "fas fa-cloud",
		HeroHeading: "Secure, Scalable Cloud Architectures Designed for Modern Enterprise Agility",
		FullDesc:    "Unlock the agility, elasticity, and cost efficiencies of cloud-native computing. Spot Solutions helps enterprises design, migrate, containerize, and optimize infrastructure across Azure, AWS, and private hosting environments.",
		Features: []string{
			"Infrastructure as Code for reproducible environments",
			"Zero-downtime database migrations and automated backup strategies",
			"Container orchestration with Docker and managed cloud services",
			"Comprehensive cloud security audits, VPC network isolation, and TLS encryption",
		},
		Benefits: []string{
			"Up to 40% reduction in unmanaged cloud resource overhead",
			"Guaranteed 99.99% uptime with automated health recovery and failover",
			"Rapid scaling to accommodate traffic spikes during peak business cycles",
		},
		Deliverables: []string{
			"Cloud Architecture Blueprints & Schematics",
			"Automated Backup & Disaster Recovery Runbooks",
			"Security Certification & SLA Documentation",
		},
	},
}

type PageHandler struct {
	cfg *config.Config
}

func NewPageHandler(cfg *config.Config) *PageHandler {
	return &PageHandler{cfg: cfg}
}

func (h *PageHandler) Home(c *gin.Context) {
	c.HTML(http.StatusOK, "home.html", gin.H{
		"Title":       "Spot Solutions | Custom Software Development & SharePoint Company",
		"Description": "Spot Solutions is a Canadian software development company with over 20 years of experience in custom software development, SharePoint, and Microsoft 365.",
		"Canonical":   h.cfg.BaseURL + "/",
		"CurrentPath": "/",
		"Services":    ServicesCatalog,
		"CSRFToken":   c.GetString("CSRFToken"),
		"AppName":     h.cfg.AppName,
		"BaseURL":     h.cfg.BaseURL,
	})
}

func (h *PageHandler) About(c *gin.Context) {
	c.HTML(http.StatusOK, "about.html", gin.H{
		"Title":       "About Us | Spot Solutions",
		"Description": "Spot Solutions builds fully customized software solutions to improve client business operations for over 20 years.",
		"Canonical":   h.cfg.BaseURL + "/about-us",
		"CurrentPath": "/about-us",
		"AppName":     h.cfg.AppName,
		"BaseURL":     h.cfg.BaseURL,
	})
}

func (h *PageHandler) Services(c *gin.Context) {
	c.HTML(http.StatusOK, "services.html", gin.H{
		"Title":       "Our Services | Spot Solutions",
		"Description": "Custom software development, SharePoint, Microsoft 365, Power Platform, Mobile Apps, and AI Advisory from Spot Solutions.",
		"Canonical":   h.cfg.BaseURL + "/services",
		"CurrentPath": "/services",
		"Services":    ServicesCatalog,
		"AppName":     h.cfg.AppName,
		"BaseURL":     h.cfg.BaseURL,
	})
}

func formatTitle(slug string) string {
	parts := strings.Split(slug, "-")
	for i, part := range parts {
		if len(part) > 0 {
			parts[i] = strings.ToUpper(part[:1]) + part[1:]
		}
	}
	return strings.Join(parts, " ")
}

func (h *PageHandler) ServiceDetail(c *gin.Context) {
	slug := c.Param("slug")
	service, exists := ServicesCatalog[slug]
	if !exists {
		// Provide high-quality fallback matching Spot Solutions context instead of 404
		title := formatTitle(slug)
		service = ServiceDetail{
			Slug:        slug,
			Title:       title,
			ShortDesc:   "Spot Solutions delivers enterprise-grade software development, consulting, and SharePoint solutions.",
			Icon:        "fas fa-code",
			HeroHeading: title + " Solutions from Spot Solutions",
			FullDesc:    "We build custom technology solutions tailored to your operational needs. As a Microsoft Solutions Partner with over 20 years of Canadian enterprise delivery, Spot Solutions provides end-to-end expertise across all project phases.",
			Features: []string{
				"Custom Architecture & Engineering tailored to your requirements",
				"Microsoft Partner Certified Standards and Best Practices",
				"Full User Training, Change Management & Adoption Support",
				"Guaranteed On-Time & On-Budget Performance Tracking",
			},
			Benefits: []string{
				"Streamlined business operations and improved employee productivity",
				"Secure and compliant Canadian data management and hosting",
				"High return on technology investment with long-term maintenance",
			},
			Deliverables: []string{
				"Production Software & Complete Source Code",
				"Architecture & Workflow Documentation",
				"User Guides & Handover Training Sessions",
			},
		}
	}

	c.HTML(http.StatusOK, "service_detail.html", gin.H{
		"Title":       service.Title + " | Spot Solutions",
		"Description": service.ShortDesc,
		"Canonical":   h.cfg.BaseURL + "/services/" + service.Slug,
		"CurrentPath": "/services/" + service.Slug,
		"Service":     service,
		"CSRFToken":   c.GetString("CSRFToken"),
		"AppName":     h.cfg.AppName,
		"BaseURL":     h.cfg.BaseURL,
	})
}

func (h *PageHandler) Industries(c *gin.Context) {
	c.HTML(http.StatusOK, "industries.html", gin.H{
		"Title":       "Industries & Solutions | Spot Solutions",
		"Description": "Discover how Spot Solutions empowers healthcare, government ministries, transportation, education, and crown corporations.",
		"Canonical":   h.cfg.BaseURL + "/industries",
		"CurrentPath": "/industries",
		"AppName":     h.cfg.AppName,
		"BaseURL":     h.cfg.BaseURL,
	})
}

func (h *PageHandler) WhyUs(c *gin.Context) {
	c.HTML(http.StatusOK, "why_us.html", gin.H{
		"Title":       "Why Choose Us | Spot Solutions",
		"Description": "Over 20 years in business, over 97% on-time delivery, and 100% Canadian engineering team.",
		"Canonical":   h.cfg.BaseURL + "/why-us",
		"CurrentPath": "/why-us",
		"AppName":     h.cfg.AppName,
		"BaseURL":     h.cfg.BaseURL,
	})
}

func (h *PageHandler) Contact(c *gin.Context) {
	c.HTML(http.StatusOK, "contact.html", gin.H{
		"Title":       "Contact Us | Spot Solutions",
		"Description": "Get in touch with Spot Solutions in Vancouver, Victoria, or Calgary for a free project consultation.",
		"Canonical":   h.cfg.BaseURL + "/contact-us",
		"CurrentPath": "/contact-us",
		"CSRFToken":   c.GetString("CSRFToken"),
		"AppName":     h.cfg.AppName,
		"BaseURL":     h.cfg.BaseURL,
	})
}

func (h *PageHandler) Portfolio(c *gin.Context) {
	c.HTML(http.StatusOK, "portfolio.html", gin.H{
		"Title":       "Portfolio & Case Studies | Spot Solutions",
		"Description": "Explore enterprise software and SharePoint case studies across British Columbia and North America.",
		"Canonical":   h.cfg.BaseURL + "/portfolio",
		"CurrentPath": "/portfolio",
		"AppName":     h.cfg.AppName,
		"BaseURL":     h.cfg.BaseURL,
	})
}

func (h *PageHandler) Careers(c *gin.Context) {
	c.HTML(http.StatusOK, "careers.html", gin.H{
		"Title":       "Careers | Spot Solutions",
		"Description": "Join our Vancouver-based team of senior software developers, SharePoint specialists, and AI consultants.",
		"Canonical":   h.cfg.BaseURL + "/careers",
		"CurrentPath": "/careers",
		"AppName":     h.cfg.AppName,
		"BaseURL":     h.cfg.BaseURL,
	})
}

func (h *PageHandler) Blog(c *gin.Context) {
	c.HTML(http.StatusOK, "blog.html", gin.H{
		"Title":       "Technology Insights & Blog | Spot Solutions",
		"Description": "Articles on software engineering, Microsoft 365, SharePoint migrations, and AI adoption strategies.",
		"Canonical":   h.cfg.BaseURL + "/blog",
		"CurrentPath": "/blog",
		"AppName":     h.cfg.AppName,
		"BaseURL":     h.cfg.BaseURL,
	})
}

func (h *PageHandler) Privacy(c *gin.Context) {
	c.HTML(http.StatusOK, "privacy.html", gin.H{
		"Title":       "Privacy Policy | Spot Solutions",
		"Description": "Read the Spot Solutions privacy policy regarding data collection, protection, and security practices.",
		"Canonical":   h.cfg.BaseURL + "/privacy",
		"CurrentPath": "/privacy",
		"AppName":     h.cfg.AppName,
		"BaseURL":     h.cfg.BaseURL,
	})
}

func (h *PageHandler) Terms(c *gin.Context) {
	c.HTML(http.StatusOK, "terms.html", gin.H{
		"Title":       "Terms and Conditions | Spot Solutions",
		"Description": "Review the terms and conditions governing the use of Spot Solutions services and website.",
		"Canonical":   h.cfg.BaseURL + "/terms",
		"CurrentPath": "/terms",
		"AppName":     h.cfg.AppName,
		"BaseURL":     h.cfg.BaseURL,
	})
}

func (h *PageHandler) NotFound(c *gin.Context) {
	c.HTML(http.StatusNotFound, "404.html", gin.H{
		"Title":       "Page Not Found (404) | Spot Solutions",
		"Description": "The page you are looking for does not exist or has been moved.",
		"CurrentPath": c.Request.URL.Path,
		"AppName":     h.cfg.AppName,
		"BaseURL":     h.cfg.BaseURL,
	})
}
