package http

import (
	"net/http"
)

// CareerHandler serves career matching endpoints.
type CareerHandler struct{}

// CareerMatch holds major/career recommendations for a single type.
type CareerMatch struct {
	TypeID   string   `json:"type_id"`
	TypeName string   `json:"type_name"`
	Majors   []string `json:"majors"`
	Careers  []string `json:"careers"`
	Avoid    []string `json:"avoid"`
	Advice   string   `json:"advice"`
}

// careerMap holds the mapping for all 25 types.
var careerMap = map[string]CareerMatch{
	"W": {
		TypeID: "W", TypeName: "Wood",
		Majors:  []string{"Computer Science", "Automation", "Architecture", "Landscape Design", "Biology", "Education", "Forestry", "Environmental Science"},
		Careers: []string{"Software Engineer", "Architect", "Biologist", "Teacher", "Environmental Consultant", "Startup Founder", "Product Manager"},
		Avoid:   []string{"Highly repetitive roles", "Rigid bureaucratic positions", "Isolated back-office work without growth path"},
		Advice:  "Your energy needs room to grow. Choose fields that reward initiative and allow your work to branch in new directions over time. Avoid roles where the path is fixed from day one.",
	},
	"F": {
		TypeID: "F", TypeName: "Fire",
		Majors:  []string{"Media Studies", "Law", "Marketing", "Performing Arts", "Design", "Psychology", "Political Science", "Journalism"},
		Careers: []string{"Journalist", "Lawyer", "Marketing Director", "Actor", "Designer", "Therapist", "Politician", "Public Speaker"},
		Avoid:   []string{"Isolated desk work without human contact", "Routine-heavy roles", "Data entry"},
		Advice:  "Your energy is expressive and influential. Choose fields where your passion and communication skills are assets, not liabilities. You thrive in roles with visibility and impact.",
	},
	"E": {
		TypeID: "E", TypeName: "Earth",
		Majors:  []string{"Clinical Medicine", "Management", "Finance", "Education", "Social Work", "Agriculture", "Nursing", "Public Health"},
		Careers: []string{"Doctor", "Manager", "Financial Advisor", "Teacher", "Social Worker", "Civil Servant", "Healthcare Administrator"},
		Avoid:   []string{"High-risk speculative roles", "Constantly changing environments without stability", "Solo entrepreneurial ventures without support"},
		Advice:  "Your energy is stable and nurturing. Choose fields where reliability and care are valued. You excel in roles that require steadiness, patience, and the ability to hold space for others.",
	},
	"M": {
		TypeID: "M", TypeName: "Metal",
		Majors:  []string{"Mathematics", "Physics", "Mechanical Engineering", "Accounting", "Law", "Medical Laboratory Science", "Materials Science", "Statistics"},
		Careers: []string{"Engineer", "Accountant", "Judge", "Data Scientist", "Quality Assurance", "Surgeon", "Research Scientist", "Actuary"},
		Avoid:   []string{"Ambiguous roles without clear metrics", "Creative fields with no structure", "Jobs that change rules frequently"},
		Advice:  "Your energy is precise and structured. Choose fields that reward rigor, accuracy, and clear thinking. You thrive where rules are consistent and excellence is measurable.",
	},
	"R": {
		TypeID: "R", TypeName: "Water",
		Majors:  []string{"Philosophy", "History", "Theoretical Physics", "Literature", "Astronomy", "Archaeology", "Music Composition", "Cognitive Science"},
		Careers: []string{"Professor", "Writer", "Research Scientist", "Composer", "Psychoanalyst", "Strategic Consultant", "Librarian", "Data Analyst"},
		Avoid:   []string{"High-pressure sales", "Superficial social roles", "Fast-paced multitasking environments"},
		Advice:  "Your energy is deep and reflective. Choose fields that reward depth over speed. You thrive where you can immerse yourself in complex problems and emerge with profound insights.",
	},
	// ... remaining 20 types follow the same pattern
	"WF": {
		TypeID: "WF", TypeName: "Wood-Fire",
		Majors:  []string{"Computer Science", "Design", "Entrepreneurship", "Digital Media", "Architecture", "Marketing"},
		Careers: []string{"Tech Entrepreneur", "UX Designer", "Creative Director", "Product Manager", "Innovation Consultant", "Venture Capitalist"},
		Avoid:   []string{"Heavily regulated industries", "Repetitive operational roles"},
		Advice:  "Wood creates, Fire expresses. You're a builder who can also sell the vision — rare and powerful. Choose fields where you can both create AND communicate.",
	},
	"FW": {
		TypeID: "FW", TypeName: "Fire-Wood",
		Majors:  []string{"Journalism", "Education", "Performing Arts", "Media Studies", "Psychology"},
		Careers: []string{"Educator", "Speaker", "Coach", "Content Creator", "Arts Director", "Community Organizer"},
		Avoid:   []string{"Isolated analytical work", "Roles with minimal human interaction"},
		Advice:  "Fire leads with expression, Wood provides creative depth. You inspire through words and presence. Choose fields where your voice reaches and moves people.",
	},
	"FE": {
		TypeID: "FE", TypeName: "Fire-Earth",
		Majors:  []string{"Clinical Medicine", "Public Health", "Hospitality Management", "Human Resources", "Social Work"},
		Careers: []string{"Physician", "HR Director", "Healthcare Leader", "Non-profit Executive", "Community Health Worker"},
		Avoid:   []string{"Purely theoretical roles", "Cold corporate environments"},
		Advice:  "Fire warms, Earth nurtures. You combine passion with practical care. Choose fields where you can serve people directly and see the impact of your work.",
	},
	"EF": {
		TypeID: "EF", TypeName: "Earth-Fire",
		Majors:  []string{"Nursing", "Education", "Public Administration", "Counseling", "Organizational Psychology"},
		Careers: []string{"Nurse Practitioner", "School Principal", "Therapist", "Government Administrator", "Non-profit Director"},
		Avoid:   []string{"Cutthroat competitive environments", "Isolated technical work"},
		Advice:  "Earth stabilizes, Fire motivates. You build institutions that last and inspire people within them. Choose fields where your steady warmth can sustain teams and communities.",
	},
	"EM": {
		TypeID: "EM", TypeName: "Earth-Metal",
		Majors:  []string{"Finance", "Accounting", "Business Administration", "Supply Chain Management", "Civil Engineering"},
		Careers: []string{"CFO", "Operations Director", "Bank Manager", "Supply Chain Lead", "Project Manager"},
		Avoid:   []string{"Speculative startups without foundation", "Unstructured creative roles"},
		Advice:  "Earth grounds, Metal structures. You build systems that are both stable and efficient. Choose fields where you can design processes that endure.",
	},
	"ME": {
		TypeID: "ME", TypeName: "Metal-Earth",
		Majors:  []string{"Industrial Engineering", "Quality Management", "Auditing", "Risk Management", "Construction Management"},
		Careers: []string{"Quality Director", "Auditor", "Risk Analyst", "Construction Manager", "Compliance Officer"},
		Avoid:   []string{"Disorganized environments", "Roles without clear standards"},
		Advice:  "Metal cuts precisely, Earth provides foundation. You create order from chaos. Choose fields where your ability to structure and standardize is valued.",
	},
	"MR": {
		TypeID: "MR", TypeName: "Metal-Water",
		Majors:  []string{"Theoretical Physics", "Computer Science", "Mathematics", "Philosophy", "Data Science"},
		Careers: []string{"Research Scientist", "Software Architect", "Quantitative Analyst", "Academic", "Systems Engineer"},
		Avoid:   []string{"Superficial roles", "Constant context-switching"},
		Advice:  "Metal gives precision, Water gives depth. You can think both rigorously and profoundly. Choose fields where deep, systematic thinking is the core requirement.",
	},
	"RM": {
		TypeID: "RM", TypeName: "Water-Metal",
		Majors:  []string{"Physics", "Philosophy", "Linguistics", "Computer Science", "Neuroscience"},
		Careers: []string{"Theoretical Physicist", "AI Researcher", "Philosopher", "Cryptographer", "Systems Theorist"},
		Avoid:   []string{"Fast-paced operational roles", "Surface-level work"},
		Advice:  "Water dives deep, Metal structures the insight. You're a deep thinker who can formalize profound ideas. Choose fields where intellectual depth is the currency.",
	},
	"RW": {
		TypeID: "RW", TypeName: "Water-Wood",
		Majors:  []string{"Biology", "Environmental Science", "Creative Writing", "Anthropology", "Marine Science"},
		Careers: []string{"Research Biologist", "Environmental Scientist", "Author", "Anthropologist", "Marine Biologist"},
		Avoid:   []string{"Rigid corporate hierarchies", "Repetitive administrative roles"},
		Advice:  "Water nourishes, Wood grows. You turn deep understanding into new growth. Choose fields where you can study complex systems and generate fresh insights.",
	},
	"WR": {
		TypeID: "WR", TypeName: "Wood-Water",
		Majors:  []string{"Environmental Engineering", "Urban Planning", "Landscape Architecture", "Education", "Psychology"},
		Careers: []string{"Urban Planner", "Landscape Architect", "Educational Psychologist", "Sustainability Consultant", "Policy Advisor"},
		Avoid:   []string{"Purely commercial roles", "Work without intellectual challenge"},
		Advice:  "Wood initiates, Water provides deep thinking. You're a visionary with substance — you don't just start things, you understand why they need to exist.",
	},
	"WE": {
		TypeID: "WE", TypeName: "Wood-Earth",
		Majors:  []string{"Business", "Agricultural Science", "Civil Engineering", "Real Estate", "Economics"},
		Careers: []string{"Business Owner", "Real Estate Developer", "Agricultural Manager", "Construction Project Lead", "Economic Analyst"},
		Avoid:   []string{"Fast-burn industries", "Ethically questionable fields"},
		Advice:  "Wood grows, Earth sustains. You build things that last. Choose fields where your long-term vision and steady execution can create enduring value.",
	},
	"EW": {
		TypeID: "EW", TypeName: "Earth-Wood",
		Majors:  []string{"Education", "Environmental Policy", "Healthcare Management", "Urban Studies", "Public Policy"},
		Careers: []string{"Policy Maker", "Healthcare Administrator", "School Administrator", "Environmental Planner", "Foundation Director"},
		Avoid:   []string{"Disruptive startups", "Roles requiring constant reinvention"},
		Advice:  "Earth provides the base, Wood provides the growth direction. You're a builder of institutions that evolve. Choose fields where steady, principled progress is the goal.",
	},
	"WM": {
		TypeID: "WM", TypeName: "Wood-Metal",
		Majors:  []string{"Engineering", "Product Design", "Robotics", "Biotechnology", "Architecture"},
		Careers: []string{"Design Engineer", "Robotics Engineer", "Biotech Researcher", "Product Designer", "Systems Architect"},
		Avoid:   []string{"Pure people-management roles", "Administrative overhead"},
		Advice:  "Wood creates, Metal refines. You can generate ideas AND execute them with precision. Choose fields where both creativity and engineering rigor are needed.",
	},
	"MW": {
		TypeID: "MW", TypeName: "Metal-Wood",
		Majors:  []string{"Mechanical Engineering", "Industrial Design", "Orthopedic Surgery", "Precision Manufacturing", "Structural Engineering"},
		Careers: []string{"Surgeon", "Industrial Designer", "Precision Engineer", "Structural Engineer", "Tool Maker"},
		Avoid:   []string{"Abstract roles without tangible output"},
		Advice:  "Metal gives precision, Wood gives direction. You build with both accuracy and vision. Choose fields where you can create things that are both beautiful and functional.",
	},
	"FM": {
		TypeID: "FM", TypeName: "Fire-Metal",
		Majors:  []string{"Law", "Forensic Science", "Investigative Journalism", "Criminal Justice", "Political Science"},
		Careers: []string{"Prosecutor", "Investigative Journalist", "Detective", "Compliance Investigator", "Political Strategist"},
		Avoid:   []string{"Passive roles", "Bystander positions"},
		Advice:  "Fire brings intensity, Metal brings incisiveness. You're sharp, passionate, and driven by justice. Choose fields where your ability to cut through noise and fight for truth is valued.",
	},
	"MF": {
		TypeID: "MF", TypeName: "Metal-Fire",
		Majors:  []string{"Surgery", "Military Science", "Competitive Sports", "Crisis Management", "Engineering Management"},
		Careers: []string{"Surgeon", "Military Officer", "Elite Athlete", "Crisis Manager", "Engineering Director"},
		Avoid:   []string{"Passive supportive roles", "Slow-moving bureaucracies"},
		Advice:  "Metal gives precision, Fire gives decisive action. You're built for high-stakes environments. Choose fields where split-second decisions and flawless execution matter.",
	},
	"FR": {
		TypeID: "FR", TypeName: "Fire-Water",
		Majors:  []string{"Creative Writing", "Film Directing", "Music", "Theology", "Depth Psychology"},
		Careers: []string{"Film Director", "Author", "Composer", "Theologian", "Psychoanalyst", "Creative Director"},
		Avoid:   []string{"Repetitive production work", "Shallow entertainment"},
		Advice:  "Fire expresses, Water plumbs the depths. Your creative work has emotional AND intellectual weight. Choose fields where your art can be both moving and meaningful.",
	},
	"RF": {
		TypeID: "RF", TypeName: "Water-Fire",
		Majors:  []string{"Philosophy", "Comparative Literature", "Art History", "Anthropology", "Musicology"},
		Careers: []string{"Art Critic", "Curator", "Cultural Theorist", "Documentary Filmmaker", "Literary Scholar"},
		Avoid:   []string{"Commercial art production", "Formulaic content creation"},
		Advice:  "Water provides depth, Fire brings it to light. You illuminate what others miss. Choose fields where you can dig deep and share what you find with an audience that needs it.",
	},
	"ER": {
		TypeID: "ER", TypeName: "Earth-Water",
		Majors:  []string{"Clinical Psychology", "Psychiatry", "Social Work", "Palliative Care", "Pastoral Counseling"},
		Careers: []string{"Clinical Psychologist", "Psychiatrist", "Hospice Worker", "Chaplain", "Grief Counselor"},
		Avoid:   []string{"Superficial helping roles", "Bureaucratic social services without client contact"},
		Advice:  "Earth holds, Water understands deeply. You're a safe harbor for people in their darkest moments. Choose fields where your capacity to hold space and understand pain can heal.",
	},
	"RE": {
		TypeID: "RE", TypeName: "Water-Earth",
		Majors:  []string{"Psychiatry", "Neurology", "Medical Research", "Anthropology", "Public Health"},
		Careers: []string{"Research Psychiatrist", "Neurologist", "Medical Anthropologist", "Epidemiologist", "Health Policy Researcher"},
		Avoid:   []string{"Purely clinical roles without research", "Shallow diagnostic work"},
		Advice:  "Water plumbs the depths, Earth provides grounded care. You combine scientific rigor with human understanding. Choose fields where you can study AND heal.",
	},
}

// GET /api/career/matches
func (h *CareerHandler) GetMatches(w http.ResponseWriter, r *http.Request) {
	typeID := r.URL.Query().Get("type")
	if typeID == "" {
		respondError(w, http.StatusBadRequest, "invalid_request", "type query parameter is required")
		return
	}

	match, ok := careerMap[typeID]
	if !ok {
		respondError(w, http.StatusNotFound, "not_found", "type not found")
		return
	}

	respondJSON(w, http.StatusOK, match)
}

// GET /api/career/types
func (h *CareerHandler) ListTypes(w http.ResponseWriter, r *http.Request) {
	type TypeSummary struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}
	var items []TypeSummary
	for _, m := range careerMap {
		items = append(items, TypeSummary{ID: m.TypeID, Name: m.TypeName})
	}
	respondList(w, items, len(items))
}
