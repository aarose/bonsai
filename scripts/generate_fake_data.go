package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

type Node struct {
	ID       string   `json:"id"`
	Content  string   `json:"content"`
	Type     string   `json:"type"`
	Parent   *string  `json:"parent,omitempty"`
	Children []string `json:"children"`
	Model    *string  `json:"model,omitempty"`
}

// Pre-generated UUIDs for consistent conversation structure
var (
	// Main conversation path
	uuid001 = "f47ac10b-58cc-4372-a567-0e02b2c3d479"
	uuid002 = "550e8400-e29b-41d4-a716-446655440001"
	uuid003 = "6ba7b810-9dad-11d1-80b4-00c04fd430c8"
	uuid004 = "6ba7b811-9dad-11d1-80b4-00c04fd430c8"
	uuid005 = "6ba7b812-9dad-11d1-80b4-00c04fd430c8"
	uuid006 = "6ba7b813-9dad-11d1-80b4-00c04fd430c8"
	uuid007 = "6ba7b814-9dad-11d1-80b4-00c04fd430c8"
	uuid008 = "6ba7b815-9dad-11d1-80b4-00c04fd430c8"

	// Branch 1: Toronto path
	uuidAlt003 = "1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d"
	uuidAlt004 = "2a3b4c5d-6e7f-8a9b-0c1d-2e3f4a5b6c7d"
	uuidAlt006 = "3a4b5c6d-7e8f-9a0b-1c2d-3e4f5a6b7c8d"
	uuidAlt007 = "4a5b6c7d-8e9f-0a1b-2c3d-4e5f6a7b8c9d"

	// Branch 2: Montreal weather path
	uuidAlt005 = "5a6b7c8d-9e0f-1a2b-3c4d-5e6f7a8b9c0d"
	uuidAlt008 = "6a7b8c9d-0e1f-2a3b-4c5d-6e7f8a9b0c1d"
	uuidAlt009 = "7a8b9c0d-1e2f-3a4b-5c6d-7e8f9a0b1c2d"
	uuidAlt010 = "8a9b0c1d-2e3f-4a5b-6c7d-8e9f0a1b2c3d"
)

// Conversation about trip planning from Ottawa with branching paths
var conversationNodes = []Node{
	// Main conversation path
	{
		ID:       uuid001,
		Content:  "I want to plan a trip by train from Ottawa to a nearby city. What are my options?",
		Type:     "user",
		Parent:   nil,
		Children: []string{uuid002},
		Model:    nil,
	},
	{
		ID:       uuid002,
		Content:  "Great choice for train travel! From Ottawa, you have several excellent options for nearby cities accessible by VIA Rail:\n\n1. **Montreal** - 2 hours, vibrant culture and cuisine\n2. **Toronto** - 4.5 hours, Canada's largest city\n3. **Kingston** - 2.5 hours, historic limestone city\n4. **Brockville** - 1.5 hours, charming riverside town\n\nHow long are you planning to stay, and what type of experience are you looking for?",
		Type:     "llm",
		Parent:   stringPtr(uuid001),
		Children: []string{uuid003, uuidAlt003}, // Added branching here
		Model:    stringPtr("claude-3.5-sonnet"),
	},
	// Main path: Food and history preference
	{
		ID:       uuid003,
		Content:  "I'm thinking 3 days over a weekend. I'd love something with good food and some historical sites to explore.",
		Type:     "user",
		Parent:   stringPtr(uuid002),
		Children: []string{uuid004},
		Model:    nil,
	},
	{
		ID:       uuid004,
		Content:  "Perfect! For a 3-day weekend trip with great food and history, I'd highly recommend **Montreal**. Here's why:\n\n**Food Scene:**\n- World-class bagels and smoked meat\n- Incredible French cuisine\n- Vibrant food markets like Jean-Talon\n\n**Historical Sites:**\n- Old Montreal (Vieux-Montréal) with cobblestone streets\n- Notre-Dame Basilica\n- Château Ramezay Museum\n\n**Weekend Logistics:**\n- Multiple daily trains from Ottawa (2-hour journey)\n- Easy to navigate on foot/metro\n- Perfect weekend getaway distance\n\nWould you like me to suggest a specific itinerary, or are you interested in learning about the other cities first?",
		Type:     "llm",
		Parent:   stringPtr(uuid003),
		Children: []string{uuid005, uuidAlt005}, // Added branching for different follow-ups
		Model:    stringPtr("claude-3.5-sonnet"),
	},
	// Main path continues: Itinerary request
	{
		ID:       uuid005,
		Content:  "Montreal sounds perfect! Could you give me a rough itinerary for the 3 days? Also, what's the best way to get around the city once I'm there?",
		Type:     "user",
		Parent:   stringPtr(uuid004),
		Children: []string{uuid006},
		Model:    nil,
	},
	{
		ID:       uuid006,
		Content:  "Excellent choice! Here's a perfect 3-day Montreal itinerary:\n\n**Day 1 - Old Montreal & Waterfront**\n- Morning: Arrive via train, check into hotel\n- Afternoon: Explore Vieux-Montréal, visit Notre-Dame Basilica\n- Evening: Dinner in Old Port, stroll along St. Lawrence River\n\n**Day 2 - Culture & Food**\n- Morning: Visit Jean-Talon Market for breakfast and local products\n- Afternoon: Explore Plateau Mont-Royal neighborhood\n- Evening: Try authentic smoked meat at Schwartz's Deli\n\n**Day 3 - Final Exploration**\n- Morning: Mount Royal Park for city views\n- Afternoon: Last-minute shopping on Sainte-Catherine Street\n- Evening: Departure train back to Ottawa\n\n**Getting Around:**\n- **STM Metro/Bus** - Excellent public transit system\n- **Walking** - Most attractions are walkable in compact areas\n- **BIXI Bikes** - Bike sharing for nice weather\n\nWould you like specific restaurant recommendations or help with train booking times?",
		Type:     "llm",
		Parent:   stringPtr(uuid005),
		Children: []string{uuid007},
		Model:    stringPtr("claude-3.5-sonnet"),
	},
	{
		ID:       uuid007,
		Content:  "This looks great! Yes, I'd love some specific restaurant recommendations. Also, should I book the train tickets in advance?",
		Type:     "user",
		Parent:   stringPtr(uuid006),
		Children: []string{uuid008},
		Model:    nil,
	},
	{
		ID:       uuid008,
		Content:  "Absolutely! Here are my top restaurant picks:\n\n**Must-Try Restaurants:**\n- **Schwartz's Deli** - Iconic smoked meat (since 1928)\n- **Joe Beef** - Upscale, creative Canadian cuisine\n- **Au Pied de Cochon** - Famous for foie gras and hearty Québécois food\n- **Olive et Gourmando** - Perfect breakfast/lunch spot in Old Montreal\n- **La Banquise** - 24/7 poutine paradise (30+ varieties!)\n- **Fairmount Bagel** or **St-Viateur Bagel** - Montreal bagel rivalry!\n\n**Train Booking:**\nYes, definitely book in advance! Here's why:\n- **Weekend routes fill up quickly**\n- **Better prices** when booking early\n- **Seat selection** available\n- **Economy vs Business class** options\n\nBook through VIA Rail's website or app. I recommend the morning departure (around 8:30 AM) to maximize your first day in Montreal.\n\nWould you like tips on where to stay, or do you have any dietary restrictions I should consider for the restaurant recommendations?",
		Type:     "llm",
		Parent:   stringPtr(uuid007),
		Children: []string{},
		Model:    stringPtr("claude-3.5-sonnet"),
	},

	// BRANCH 1: Alternative response to initial options - choosing Toronto instead
	{
		ID:       uuidAlt003,
		Content:  "Actually, I'm more interested in Toronto. What would a 3-day Toronto trip look like? I'm particularly interested in museums and cultural experiences.",
		Type:     "user",
		Parent:   stringPtr(uuid002),
		Children: []string{uuidAlt004},
		Model:    nil,
	},
	{
		ID:       uuidAlt004,
		Content:  "Excellent choice! Toronto offers incredible museums and cultural experiences. Here's what I'd recommend for 3 days:\n\n**Museums & Culture:**\n- **Royal Ontario Museum (ROM)** - Natural history and world cultures\n- **Art Gallery of Ontario** - Canadian and international art\n- **CN Tower & EdgeWalk** - Iconic views and thrills\n- **Distillery District** - Historic cobblestone streets and galleries\n- **Casa Loma** - Gothic Revival castle and gardens\n\n**Neighborhoods to Explore:**\n- **Kensington Market** - Eclectic, bohemian atmosphere\n- **Queen Street West** - Trendy shops and street art\n- **Harbourfront** - Waterfront walks and events\n\n**Practical Info:**\n- Train arrives at Union Station (downtown core)\n- Excellent TTC subway system\n- 4.5-hour journey from Ottawa\n\nWould you like a day-by-day breakdown, or are you curious about specific aspects like food scene or accommodation?",
		Type:     "llm",
		Parent:   stringPtr(uuidAlt003),
		Children: []string{uuidAlt006},
		Model:    stringPtr("claude-3.5-sonnet"),
	},
	{
		ID:       uuidAlt006,
		Content:  "This sounds amazing! Could you give me that day-by-day breakdown? Also, what's the weather usually like in Toronto during spring?",
		Type:     "user",
		Parent:   stringPtr(uuidAlt004),
		Children: []string{uuidAlt007},
		Model:    nil,
	},
	{
		ID:       uuidAlt007,
		Content:  "Perfect! Here's your detailed Toronto itinerary with spring weather info:\n\n**Day 1 - Downtown Core**\n- Morning: Arrive, check in, visit CN Tower\n- Afternoon: Royal Ontario Museum\n- Evening: Dinner in Entertainment District\n\n**Day 2 - Culture & Neighborhoods**\n- Morning: Art Gallery of Ontario\n- Afternoon: Explore Kensington Market and Queen West\n- Evening: Distillery District for dinner and drinks\n\n**Day 3 - Final Adventures**\n- Morning: Casa Loma castle\n- Afternoon: Harbourfront walk or last-minute shopping\n- Evening: Departure from Union Station\n\n**Spring Weather (March-May):**\n- **March**: Cool, 5-15°C, bring layers and waterproof jacket\n- **April**: Mild, 10-18°C, perfect walking weather\n- **May**: Pleasant, 15-22°C, ideal for outdoor activities\n- **General**: Pack layers, light rain jacket, comfortable walking shoes\n\nSpring is actually one of the best times to visit - fewer crowds, mild weather, and the city awakening from winter!",
		Type:     "llm",
		Parent:   stringPtr(uuidAlt006),
		Children: []string{},
		Model:    stringPtr("claude-3.5-sonnet"),
	},

	// BRANCH 2: Climate inquiry branch from Montreal recommendation
	{
		ID:       uuidAlt005,
		Content:  "Before deciding on Montreal, what's the weather like there in early spring? I want to make sure I pack appropriately.",
		Type:     "user",
		Parent:   stringPtr(uuid004),
		Children: []string{uuidAlt008},
		Model:    nil,
	},
	{
		ID:       uuidAlt008,
		Content:  "Great question! Montreal's early spring weather is quite variable, so packing right is key:\n\n**Early Spring in Montreal (March-April):**\n- **March**: Still chilly, -2°C to 8°C, snow possible\n- **April**: Milder, 4°C to 15°C, rain more common than snow\n- **Typical conditions**: Cool mornings, warmer afternoons\n\n**What to Pack:**\n- **Layered clothing** - temperatures change throughout the day\n- **Waterproof jacket** - spring showers are common\n- **Comfortable waterproof boots** - for walking on potentially wet cobblestones\n- **Light sweater or fleece** - for cooler evenings\n- **Scarf and light gloves** - especially for early March\n\n**Silver Lining:**\n- Fewer tourists, so popular spots are less crowded\n- Hotels often have better rates\n- Indoor attractions (museums, markets) are perfect for cooler days\n- Hot chocolate and warm café culture is at its peak!\n\nThe variable weather actually adds to Montreal's charm - cozy cafés become even more inviting! Would you like that detailed itinerary now, keeping the weather in mind?",
		Type:     "llm",
		Parent:   stringPtr(uuidAlt005),
		Children: []string{uuidAlt009},
		Model:    stringPtr("claude-3.5-sonnet"),
	},
	{
		ID:       uuidAlt009,
		Content:  "That's really helpful! Yes, I'd love the itinerary adapted for the spring weather. Maybe with more indoor backup options?",
		Type:     "user",
		Parent:   stringPtr(uuidAlt008),
		Children: []string{uuidAlt010},
		Model:    nil,
	},
	{
		ID:       uuidAlt010,
		Content:  "Absolutely! Here's a weather-smart Montreal itinerary with indoor alternatives:\n\n**Day 1 - Flexible Old Montreal**\n- **Morning**: Arrive, check in, warm up with coffee\n- **Indoor option**: Notre-Dame Basilica + Pointe-à-Callière Museum\n- **Outdoor option**: Old Montreal walk (if weather permits)\n- **Evening**: Cozy dinner at Modavie or similar indoor venue\n\n**Day 2 - Indoor Culture + Covered Markets**\n- **Morning**: **Jean-Talon Market** (partially covered, perfect for any weather)\n- **Afternoon**: **Montreal Museum of Fine Arts** or **McCord Museum**\n- **Backup**: **Underground City** exploration if weather turns bad\n- **Evening**: **Plateau indoor dining** - plenty of cozy bistros\n\n**Day 3 - Adaptable Final Day**\n- **Good weather**: Mount Royal Park + outdoor exploration\n- **Poor weather**: **Biodome** or **Olympic Stadium** tour\n- **Always good**: **Sainte-Catherine Street** shopping (mostly covered)\n- **Before departure**: Warm meal at Central Station\n\n**Pro tip**: Montreal's Underground City connects major downtown areas, so you can stay warm while exploring if needed!",
		Type:     "llm",
		Parent:   stringPtr(uuidAlt009),
		Children: []string{},
		Model:    stringPtr("claude-3.5-sonnet"),
	},
}

func stringPtr(s string) *string {
	return &s
}

func main() {
	var dbPath string
	
	// Check if custom path provided via command line
	if len(os.Args) > 1 {
		dbPath = os.Args[1]
	} else {
		// Use same path as CLI tool: ~/.bonsai/bonsai.db
		homeDir, err := os.UserHomeDir()
		if err != nil {
			log.Fatalf("Failed to get home directory: %v", err)
		}
		dbPath = filepath.Join(homeDir, ".bonsai", "bonsai.db")
	}

	// Ensure the directory exists
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		log.Fatalf("Failed to create database directory: %v", err)
	}

	// Open database connection
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Test connection
	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	// Create table if it doesn't exist
	createNodeTable := `
	CREATE TABLE IF NOT EXISTS Node (
		id TEXT PRIMARY KEY,
		content TEXT NOT NULL,
		type TEXT NOT NULL CHECK (type IN ('user', 'llm')),
		parent TEXT,
		children TEXT DEFAULT '[]',
		model TEXT
	);`

	if _, err := db.Exec(createNodeTable); err != nil {
		log.Fatalf("Failed to create Node table: %v", err)
	}

	// Clear existing data (optional - comment out if you want to keep existing data)
	if _, err := db.Exec("DELETE FROM Node"); err != nil {
		log.Fatalf("Failed to clear existing data: %v", err)
	}

	// Insert conversation nodes
	insertSQL := `
	INSERT INTO Node (id, content, type, parent, children, model)
	VALUES (?, ?, ?, ?, ?, ?)`

	stmt, err := db.Prepare(insertSQL)
	if err != nil {
		log.Fatalf("Failed to prepare statement: %v", err)
	}
	defer stmt.Close()

	for _, node := range conversationNodes {
		// Convert children slice to JSON string
		childrenJSON, err := json.Marshal(node.Children)
		if err != nil {
			log.Fatalf("Failed to marshal children for node %s: %v", node.ID, err)
		}

		// Execute insert
		_, err = stmt.Exec(
			node.ID,
			node.Content,
			node.Type,
			node.Parent,
			string(childrenJSON),
			node.Model,
		)
		if err != nil {
			log.Fatalf("Failed to insert node %s: %v", node.ID, err)
		}

		fmt.Printf("Inserted node: %s (%s)\n", node.ID, node.Type)
	}

	fmt.Printf("\nSuccessfully generated fake conversation data!\n")
	fmt.Printf("Database: %s\n", dbPath)
	fmt.Printf("Total nodes: %d\n", len(conversationNodes))
	fmt.Println("\nConversation paths created:")
	fmt.Println("- Main path: Ottawa trip planning → Montreal recommendation → Detailed itinerary")
	fmt.Println("- Branch 1: Toronto alternative with cultural focus and weather info")
	fmt.Println("- Branch 2: Montreal weather inquiry leading to adapted itinerary")
}
