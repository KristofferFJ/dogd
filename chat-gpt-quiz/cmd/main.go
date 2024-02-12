package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/sashabaranov/go-openai"
	"math/rand"
	"net/http"
	"os"
	"strings"
)

const baseQuery = "Please describe the following magic card to me, without using its name. This is for a quiz, so please be brief. "

type App struct {
	client *openai.Client
}

func main() {
	err := godotenv.Load()
	if err != nil {
		return
	}
	var key = os.Getenv("OPENAI_API")
	var client = openai.NewClient(key)
	app := &App{client: client}
	if err := app.run(); err != nil {
		fmt.Printf("%s\n", err)
		os.Exit(1)
	}
}

func (app *App) run() error {
	router := mux.NewRouter()
	router.HandleFunc("/1", app.quizOneHandler)
	router.HandleFunc("/2", app.quizTwoHandler)
	router.HandleFunc("/3", app.quizThreeHandler)
	router.HandleFunc("/4", app.quizFourHandler)
	router.HandleFunc("/5", app.quizFiveHandler)
	router.HandleFunc("/scryfall", app.testScryfallHandler)
	router.Use(corsMiddleware)
	err := http.ListenAndServe(":1337", router)
	if err != nil {
		return err
	}
	return nil
}

// corsMiddleware adds CORS headers to the response
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*")                                // Allow any origin
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE") // Allowed methods
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

		// Check if the request is for the OPTIONS method (pre-flight request)
		if r.Method == "OPTIONS" {
			// Respond with allowed methods
			w.WriteHeader(http.StatusOK)
			return
		}

		// Pass down the request to the next handler
		next.ServeHTTP(w, r)
	})
}

func (app *App) requestPictureDescription(query string, url string) string {
	resp, err := app.client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model:     openai.GPT4VisionPreview,
			MaxTokens: 300,
			Messages: []openai.ChatCompletionMessage{
				{
					Role: openai.ChatMessageRoleUser,
					MultiContent: []openai.ChatMessagePart{
						{
							Type: openai.ChatMessagePartTypeText,
							Text: query,
						},
						{
							Type: openai.ChatMessagePartTypeImageURL,
							ImageURL: &openai.ChatMessageImageURL{
								URL:    url,
								Detail: openai.ImageURLDetailAuto,
							},
						},
					},
				},
			},
		},
	)

	if err != nil {
		panic(err)
	}

	return resp.Choices[0].Message.Content
}

func (app *App) quizOneHandler(w http.ResponseWriter, _ *http.Request) {
	card := getRandomCubeCard()
	scryfall := getScryfallInformation(card)
	response := app.requestPictureDescription(baseQuery+card, scryfall.ImageUris.Large)
	JsonResponse(w, Response{response, scryfall.ImageUris.Large})
}

func (app *App) quizTwoHandler(w http.ResponseWriter, _ *http.Request) {
	card := getRandomCubeCard()
	scryfall := getScryfallInformation(card)
	response := app.requestPictureDescription(baseQuery+"Do not mention power / toughness. "+card, scryfall.ImageUris.Large)
	JsonResponse(w, Response{response, scryfall.ImageUris.Large})
}

func (app *App) quizThreeHandler(w http.ResponseWriter, _ *http.Request) {
	card := getRandomCubeCard()
	scryfall := getScryfallInformation(card)
	response := app.requestPictureDescription(baseQuery+"Ignore the artwork. "+card, scryfall.ImageUris.Large)
	JsonResponse(w, Response{response, scryfall.ImageUris.Large})
}

func (app *App) quizFourHandler(w http.ResponseWriter, _ *http.Request) {
	card := getRandomCubeCard()
	scryfall := getScryfallInformation(card)
	response := app.requestPictureDescription(baseQuery+"Describe only the artwork. "+card, scryfall.ImageUris.Large)
	JsonResponse(w, Response{response, scryfall.ImageUris.Large})
}

func (app *App) quizFiveHandler(w http.ResponseWriter, _ *http.Request) {
	card := getRandomCubeCard()
	scryfall := getScryfallInformation(card)
	response := app.requestPictureDescription(baseQuery+"Describe only the artwork, but rap the answer. "+card, scryfall.ImageUris.Large)
	JsonResponse(w, Response{response, scryfall.ImageUris.Large})
}

func (app *App) testScryfallHandler(w http.ResponseWriter, _ *http.Request) {
	card := getRandomCubeCard()
	scryfall := getScryfallInformation(card)
	JsonResponse(w, Response{scryfall.Name, scryfall.ImageUris.Large})
}

type ImageUris struct {
	Small      string `json:"small"`
	Normal     string `json:"normal"`
	Large      string `json:"large"`
	Png        string `json:"png"`
	ArtCrop    string `json:"art_crop"`
	BorderCrop string `json:"border_crop"`
}

type ScryfallCard struct {
	Object     string    `json:"object"`
	Id         string    `json:"id"`
	OracleId   string    `json:"oracle_id"`
	Name       string    `json:"name"`
	ReleasedAt string    `json:"released_at"`
	ImageUris  ImageUris `json:"image_uris"`
	ManaCost   string    `json:"mana_cost"`
	Cmc        float64   `json:"cmc"`
	TypeLine   string    `json:"type_line"`
	OracleText string    `json:"oracle_text"`
}

func getScryfallInformation(name string) ScryfallCard {
	res, err := http.Get("https://api.scryfall.com/cards/named?order=released&dir=asc&exact=" + name)
	if err != nil {
		fmt.Println(err)
		panic(err)
	}
	var card ScryfallCard
	err = json.NewDecoder(res.Body).Decode(&card)
	if err != nil {
		panic(err)
	}
	return card
}

func getRandomCubeCard() string {
	cards := strings.Split(CUBE, "\n")
	return cards[rand.Intn(len(cards))]
}

type Response struct {
	Text string
	Url  string
}

func JsonResponse(w http.ResponseWriter, data Response) {
	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(data)
	if err != nil {
		return
	}
}

var CUBE = `Champion of the Parish
Dauntless Bodyguard
Esper Sentinel
Giant Killer
Giver of Runes
Kytheon, Hero of Akros
Mother of Runes
Student of Warfare
Thraben Inspector
Usher of the Fallen
Adanto Vanguard
Charming Prince
Luminarch Aspirant
Samwise the Stouthearted
Scholar of New Horizons
Selfless Spirit
Stoneforge Mystic
Thalia, Guardian of Thraben
Thalia's Lieutenant
Adeline, Resplendent Cathar
Archon of Emeria
Blade Splicer
Elite Spellbinder
Flickerwisp
Loran of the Third Path
Porcelain Legionnaire
Ranger-Captain of Eos
Recruiter of the Guard
Skyclave Apparition
Halvar, God of Battle
Hero of Bladehold
Serra Paragon
Angel of Invention
Cloudgoat Ranger
Solitude
Eagles of the North
Steel Seraph
The Wandering Emperor
Enlightened Tutor
Ephemerate
Path to Exile
Secure the Wastes
Swords to Plowshares
Get Lost
Reprieve
Unexpectedly Absent
Balance
Council's Judgment
Sevinne's Reclamation
Wrath of God
Portable Hole
Staff of the Storyteller
Maul of the Skyclaves
Touch the Spirit Realm
Felidar Retreat
Parallax Wave
Leyline Binding
Enclave Cryptologist
Faerie Mastermind
Hypnotic Sprite
Jace, Vryn's Prodigy
Ledger Shredder
Malcolm, Alluring Scoundrel
Phantasmal Image
Snapcaster Mage
Barrin, Tolarian Archmage
Brazen Borrower
Champion of Wits
Chrome Host Seedshark
Deceiver Exarch
Emry, Lurker of the Loch
Hullbreacher
Man-o'-War
Master of Etherium
Pestermite
Sai, Master Thopterist
Spellseeker
Trinket Mage
Vendilion Clique
Displacer Kitten
Dungeon Geists
Phyrexian Metamorph
Subtlety
Thassa, Deep-Dwelling
Urza, Lord High Artificer
Whirler Rogue
Iymrith, Desert Doom
Mulldrifter
Stormwing Entity
Ethereal Forager
Narset, Parter of Veils
Jace, the Mind Sculptor
Tezzeret the Seeker
Ancestral Recall
Brainstorm
Consider
Force Spike
High Tide
Mystical Tutor
Stern Scolding
Syncopate
Brain Freeze
Counterspell
Daze
Mana Drain
Mana Leak
Memory Lapse
Miscalculation
Remand
Thirst for Knowledge
Cryptic Command
Fact or Fiction
Force of Will
Mystic Confluence
Sublime Epiphany
Dig Through Time
Ponder
Preordain
Chart a Course
Time Walk
Compulsive Research
Timetwister
Tinker
Lórien Revealed
Time Warp
Echo of Eons
Time Spiral
Upheaval
Treasure Cruise
Search for Azcanta
Opposition
The Antiquities War
Shark Typhoon
Bloodsoaked Champion
Carrion Feeder
Champion of the Perished
Concealing Curtains
Cryptbreaker
Dread Wanderer
Evolved Sleeper
Gravecrawler
Knight of the Ebon Legion
Blood Artist
Bloodghast
Dark Confidant
Dauthi Voidwalker
Deep-Cavern Bat
Jadar, Ghoulcaller of Nephalia
Mesmeric Fiend
Orcish Bowmasters
Priest of Forgotten Gods
Skyclave Shade
Tainted Adversary
Tourach, Dread Cantor
Zulaport Cutthroat
Graveyard Trespasser
Grim Haruspex
Midnight Reaper
Morbid Opportunist
Ophiomancer
Preacher of the Schism
Woe Strider
Grief
Phyrexian Obliterator
Rankle, Master of Pranks
Sheoldred, the Apocalypse
Yawgmoth, Thran Physician
Marionette Master
Massacre Wurm
Tasigur, the Golden Fang
Troll of Khazad-dûm
Archon of Cruelty
Griselbrand
Liliana of the Veil
Liliana, the Last Hope
Dark Ritual
Entomb
Fatal Push
Vampiric Tutor
Village Rites
Bitter Triumph
Goryo's Vengeance
Sheoldred's Edict
Dismember
Snuff Out
Bloodchief's Thirst
Bone Shards
Bone Shards
Duress
Imperial Seal
Inquisition of Kozilek
Mind Twist
Reanimate
Thoughtseize
Unearth
Collective Brutality
Demonic Tutor
Exhume
Hymn to Tourach
Night's Whisper
Night's Whisper
Sinkhole
Toxic Deluge
Yawgmoth's Will
Damnation
Tendrils of Agony
Living Death
Bolas's Citadel
Animate Dead
Bitterblossom
Dance of the Dead
The Meathook Massacre
Bastion of Remembrance
Necromancy
Recurring Nightmare
Bomat Courier
Dragon's Rage Channeler
Falkenrath Pit Fighter
Grim Lavamancer
Monastery Swiftspear
Ragavan, Nimble Pilferer
Bloodthirsty Adversary
Dockside Extortionist
Embereth Shieldbreaker
Flametongue Yearling
Inti, Seneschal of the Sun
Kari Zev, Skyship Raider
Magda, Brazen Outlaw
Robber of the Rich
Anax, Hardened in the Forge
Bonecrusher Giant
Breya's Apprentice
Broadside Bombardiers
Feldon of the Third Path
Goblin Rabblemaster
Gut, True Soul Zealot
Hanweir Garrison
Imperial Recruiter
Laelia, the Blade Reforged
Reckless Stormseeker
Seasoned Pyromancer
Flametongue Kavu
Hellrider
Hero of Oxid Ridge
Pia and Kiran Nalaar
Rampaging Raptor
Fury
Goldspan Dragon
Goldspan Dragon
Kiki-Jiki, Mirror Breaker
Siege-Gang Commander
Zealous Conscripts
Inferno Titan
Oliphaunt
Ruin Grinder
Trumpeting Carnosaur
Chandra, Torch of Defiance
Daretti, Scrap Savant
Burst Lightning
Lightning Bolt
Unholy Heat
Abrade
Incinerate
Magma Jet
Searing Blaze
Char
Seething Song
Mine Collapse
Through the Breach
Fireblast
Chain Lightning
Faithless Looting
Firebolt
Reckless Charge
Shatterskull Smashing
Tribal Flames
Arc Lightning
Wheel of Fortune
Fiery Confluence
Pyrite Spellbomb
Bitter Reunion
Goblin Bombardment
Underworld Breach
Fable of the Mirror-Breaker
Fires of Invention
Sneak Attack
Splinter Twin
Arbor Elf
Birds of Paradise
Delighted Halfling
Elvish Reclaimer
Experiment One
Gilded Goose
Hexdrinker
Ignoble Hierarch
Llanowar Elves
Noble Hierarch
Pelt Collector
Dragonsguard Elite
Elvish Visionary
Lotus Cobra
Nishoba Brawler
Rofellos, Llanowar Emissary
Sakura-Tribe Elder
Scavenging Ooze
Sylvan Advocate
Sylvan Caryatid
Tough Cookie
Voracious Hydra
Augur of Autumn
Briarbridge Tracker
Endurance
Eternal Witness
Lovestruck Beast
Ramunap Excavator
Rishkar, Peema Renegade
Sentinel of the Nameless City
Tireless Provisioner
Tireless Tracker
Questing Beast
Ulvenwald Oddity
Vengevine
World Shaper
Elder Gargaroth
Titania, Protector of Argoth
Carnage Tyrant
Generous Ent
Primeval Titan
Avenger of Zendikar
Hornet Queen
Craterhoof Behemoth
Ghalta, Primal Hunger
Woodfall Primus
Garruk Wildspeaker
Nissa, Who Shakes the World
Nissa, Ascended Animist
Blossoming Defense
Crop Rotation
Harrow
Collected Company
Abundant Harvest
Green Sun's Zenith
Pest Infestation
Edge of Autumn
Life from the Loam
Regrowth
Winding Way
Dryad's Revival
Natural Order
Plow Under
Birthing Pod
Esika's Chariot
The Great Henge
Exploration
Fastbond
Rancor
Utopia Sprawl
Oath of Druids
Survival of the Fittest
Sylvan Library
Court of Garenbrig
Geist of Saint Traft
Soulherder
Spell Queller
Teferi, Time Raveler
Fractured Identity
Teferi, Hero of Dominaria
Baleful Strix
Kaito Shizuki
Thief of Sanity
Hostage Taker
Notion Thief
Fallen Shinobi
Dreadbore
Kroxa, Titan of Death's Hunger
Daretti, Ingenious Iconoclast
Fire Covenant
Judith, the Scourge Diva
Kolaghan's Command
Mayhem Devil
Decadent Dragon
Falkenrath Aristocrat
Murderous Redcap
Olivia Voldaren
Orcish Lumberjack
Ancient Grudge
Wrenn and Six
Radha, Heart of Keld
Bloodbraid Elf
Huntmaster of the Fells
Minsc & Boo, Timeless Heroes
Escape to the Wilds
Dragonlord Atarka
Etali, Primal Conqueror
Kitchen Finks
Knight of Autumn
Knight of the Reliquary
Cruel Celebrant
Damn
Tidehollow Sculler
Anguished Unmaking
Lingering Souls
Lurrus of the Dream-Den
Vindicate
Expressive Iteration
Izzet Charm
Third Path Iconoclast
Dack Fayden
Saheeli, Sublime Artificer
Fire // Ice
Niv-Mizzet, Parun
Abrupt Decay
Fiend Artisan
Grist, the Hunger Tide
Maelstrom Pulse
Binding the Old Gods
The Gitrog Monster
Figure of Destiny
Forth Eorlingas!
Ajani Vengeant
Showdown of the Skalds
Winota, Joiner of Forces
Growth Spiral
Hydroid Krasis
Uro, Titan of Nature's Wrath
Mystic Snake
Nicol Bolas, God-Pharaoh
Korvold, Fae-Cursed King
Warden of the First Tree
Siege Rhino
Leovold, Emissary of Trest
Omnath, Locus of Creation
Atraxa, Grand Unifier
Niv-Mizzet Reborn
Hangarback Walker
Stonecoil Serpent
Walking Ballista
Signal Pest
Phyrexian Revoker
Syr Ginger, the Meal Ender
Lodestone Golem
Golos, Tireless Pilgrim
Wurmcoil Engine
Myr Battlesphere
Emrakul, the Aeons Torn
Sundering Titan
Triplicate Titan
Ulamog, the Ceaseless Hunger
Karn, Scion of Urza
Black Lotus
Chrome Mox
Engineered Explosives
Lion's Eye Diamond
Lotus Petal
Mana Crypt
Mishra's Bauble
Mox Diamond
Mox Emerald
Mox Jet
Mox Opal
Mox Pearl
Mox Ruby
Mox Sapphire
Spellbook
Urza's Bauble
Zuran Orb
Candelabra of Tawnos
Chromatic Star
Currency Converter
Eater of Virtue
Mana Vault
Retrofitter Foundry
Shadowspear
Skullclamp
Sol Ring
Agatha's Soul Cauldron
Ankh of Mishra
Grim Monolith
Mind Stone
Pentad Prism
Reckoner Bankbuster
Smuggler's Copter
Sword of the Meek
Umezawa's Jitte
Winter Orb
Basalt Monolith
Coalition Relic
Crucible of Worlds
Grafted Wargear
Mimic Vat
Nettlecyst
Palantír of Orthanc
Vedalken Shackles
Mystic Forge
Teferi's Puzzle Box
The One Ring
Batterskull
Memory Jar
The Mightstone and Weakstone
Coveted Jewel
Portal to Phyrexia
Azorius Signet
Dimir Signet
Rakdos Signet
Gruul Signet
Selesnya Signet
Orzhov Signet
Izzet Signet
Golgari Signet
Boros Signet
Simic Signet
Celestial Colonnade
Flooded Strand
Hallowed Fountain
Temple of Enlightenment
Tundra
Creeping Tar Pit
Polluted Delta
Temple of Deceit
Underground Sea
Watery Grave
Badlands
Blood Crypt
Bloodstained Mire
Lavaclaw Reaches
Sulfurous Springs
Raging Ravine
Stomping Ground
Taiga
Wooded Foothills
Savannah
Stirring Wildwood
Temple Garden
Windswept Heath
Godless Shrine
Marsh Flats
Scrubland
Shambling Vent
Silent Clearing
Scalding Tarn
Steam Vents
Volcanic Island
Wandering Fumarole
Bayou
Hissing Quagmire
Nurturing Peatland
Overgrown Tomb
Verdant Catacombs
Arid Mesa
Needle Spires
Plateau
Sacred Foundry
Sunbaked Canyon
Breeding Pool
Lumbering Falls
Misty Rainforest
Tropical Island
Spara's Headquarters
Raffine's Tower
Xander's Lounge
Ziatora's Proving Ground
Jetmir's Garden
Savai Triome
Ketria Triome
Indatha Triome
Raugrin Triome
Zagoth Triome
Ancient Tomb
Dark Depths
Evolving Wilds
Fabled Passage
Field of the Dead
Gaea's Cradle
Karakas
Library of Alexandria
Mishra's Factory
Mishra's Workshop
Prismatic Vista
Rishadan Port
Shelldock Isle
Strip Mine
Terramorphic Expanse
Thespian's Stage
Tolarian Academy
Urza's Saga
Wasteland
Westvale Abbey`
