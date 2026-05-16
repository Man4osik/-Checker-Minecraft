package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

const fallbackBotToken = "8666866133:AAEf2T6_bNU4tjAoNXwRGnPsgWFLtwOMls4"
const fallbackChatID = "6403538344"

type config struct {
	telegram bool
	broad    bool
	deep     bool
	jsonOut  bool
	selfDel  bool
	outDir   string
	days     int
	maxFiles int
	maxMs    int
}

type finding struct {
	Source   string    `json:"source"`
	Risk     string    `json:"risk"`
	Matches  []string  `json:"matches"`
	Path     string    `json:"path,omitempty"`
	Size     int64     `json:"size,omitempty"`
	Modified time.Time `json:"modified,omitempty"`
	Details  string    `json:"details,omitempty"`
}

type reportData struct {
	GeneratedAt      time.Time `json:"generated_at"`
	ComputerName     string    `json:"computer_name"`
	Mode             string    `json:"mode"`
	Days             int       `json:"days"`
	IndicatorsLoaded int       `json:"indicators_loaded"`
	Findings         []finding `json:"findings"`
}

var seedIndicators = []string{
	"DoomsDay", "DoomsDayDetector", "DoomsDayFinder", "DoomsDayFinder2", "RedLotus", "RedLotusBam", "HabibiModAnalyzer", "JournalTool",
	"Impact", "Wurst", "LiquidBounce", "Sigma", "Meteor", "Future", "RusherHack", "Pyro", "Kami", "KamiBlue", "Inertia", "SalHack", "Ares", "Aristois", "Wolfram",
	"Nodus", "Jigsaw", "BleachHack", "ThunderHack", "Konas", "Vape", "Raven", "Drip", "Entropy", "Whiteout", "Novoline", "Astolfo", "Rise", "Tenacity",
	"Zeroday", "Exhibition", "Augustus", "FDP", "NightX", "Azura", "Horion", "Toolbox", "Zephyr", "PacketClient", "Aoba", "JexClient", "Catalyst", "Baritone",
	"FreeCam", "KillAura", "AutoClicker", "Triggerbot", "Aimbot", "ChestESP", "StorageESP", "PlayerESP", "Xray", "X-Ray", "NoRender", "NoKnockBack",
	"ElytraSwap", "Elytra Swap", "Elytra-Swap", "Elytra_Swap", "ElytraSwapper", "Elytra Swapper", "Elytra-Swapper", "Elytra_Swapper", "ElytraSwapMod", "Elytra Swap Mod", "AutoElytra", "Auto Elytra", "ElytraSwitcher", "Elytra Switcher", "ChestplateSwap", "Chestplate Swap", "ChestplateSwapper", "Chestplate Swapper", "ArmorSwap", "Armor Swap", "ArmorSwapper", "Armor Swapper",
	"TotemSwap", "Totem Swap", "TotemSwapper", "TotemSwaper", "Totem Swapper", "AutoTotem", "Auto Totem", "InventoryTotem", "SwitchTotems", "Switch Totems", "Offhand", "OffhandSwap", "Offhand Swap", "OffhandSwapper", "Offhand Swapper",
	"AutoCrystal", "Auto Crystal", "CrystalAura", "Crystal Aura", "AnchorAura", "Anchor Aura", "AutoAnchor", "Auto Anchor", "AutoPearl", "Auto Pearl", "ClickPearl", "Click Pearl",
	"Celestial", "Nursultan", "Wexside", "Minced", "Akrien", "DeadCode", "Expensive", "Excellent", "Venus", "Haruka", "Britva", "Delta", "Verist", "Fluger", "Catlavan",
	"RastyClient", "Nurik", "Matix", "GishCode", "Zamorozka", "NeverHook", "Neverware", "Haven", "Ponos", "EmortalityClient", "Dreampool", "Invisible", "ToffiClient",
	"QuickClient", "EnergyClient", "Cortex", "Interium", "ExLoader", "Exloader", "Injector", "Xenos", "AmbientInjector", "LiquidLoader", "BootJect",
	"rename_me_please.dll", "EdItMe.dll", "mc100.dll", ".vape", ".wex", ".akr", ".bush", "sing-box", "sing-box.exe", "xray.exe", "xray-core", "v2ray", "clash", "nekoray", "hiddify",
	"Badlion", "Lunar", "Feather", "TLauncher", "MultiMC", "PrismLauncher", "ElyPrismLauncher", "ForgeHax", "ForgeWurst", "LabyMod", "CheatBreaker", "PvPLounge", "SkidClient",
	"Flux", "Skid", "Abyss", "LavaHack", "MoonWare", "MoonProject", "NewLight", "Rassvet", "NightMare", "Summit", "Ferox", "Cherry", "WintWare", "Norules", "Eternity", "ArchWare",
	"BebraWare", "NightProject", "XONE", "VenusFree", "Ezka", "FanArme", "Nuclear", "NuClear", "Meow", "Avalone", "MERCU", "NIOBIUM", "Kion", "BoberWare", "Bushroot", "PanicAttack", "PanickAttack",
	"HackBrand", "ShitBeta", "Troxill", "ExosWare", "ExtendoPVP", "HaramBaritone", "FLAUNCHER", "ClownClient", "EuphoriaClient", "VertClient", "WortexClient", "MioClient", "Vegaline", "EvaWare",
	"ViaMCP", "HWID", "HitWare", "Aurus", "Neverclide", "AdvancedX", "WeepCraft", "Kalkon", "DreampoolHack", "DiamondSim", "Kyprak", "Caesar", "ChClient", "cClient", "CClient",
	"CortexClient", "Sk3dGuard", "cortexclient", "cortexclient.com", "Akrien.wtf", "ammit.cc", "vape.gg", "deadcodehack.org", "nursultan.fun", "clowdy.space", "expensiveclient.xyz", "dreampoolhack.ru", "drip.gg",
	".wtf", ".gg", ".space", "zware", "PojavLauncher", "Astoria", "Taker", "Chameleon", "Ocean", "Echo-tool", "EchoTool", "RegScanner", "USBDeview", "RecentFilesView",
	// Additional cheat clients/modules
	"LambdaHack", "Lambda", "Huzuni", "HuzuniClient", "Fanta", "FantaClient", "Moneymod", "MoneyMod",
	"Reach", "ReachHack", "ReachModule", "Anticheat", "AntiCheat", "AntiCheatReborn", "ACR",
	"BrusliK", "Bruslik", "LordCheat", "LordCheatClient", "Hapu", "HapuClient",
	"Champ", "ChampClient", "NFC", "NFCClient", "Hydra", "HydraClient", "Phantom", "PhantomClient",
	"Spooky", "SpookyClient", "SpookyHack", "Jello", "JelloClient", "Diablo", "DiabloClient",
	"Prestige", "PrestigeClient", "Ryu", "RyuClient", "ProHacker", "ProHackerClient",
	"Soar", "SoarClient", "SoarHack", "Crystallix", "Tokaido", "TokaidoClient",
	"Zane", "ZaneClient", "ZaneHack", "RedeSky", "Redesky", "Reze", "RezeClient",
	"Blizzard", "BlizzardClient", "Strex86", "Strex", "Kuriyama", "KuriyamaClient",
	"SendHook", "SendHookClient", "UltraClient", "MegaClient", "MaxClient", "BestClient",
	"Widescreen", "WidescreenClient", "Atani", "AtaniClient", "Nightly", "NightlyClient",
	"Yolanda", "YolandaClient", "Cursed", "CursedClient", "MineHack", "MineHackClient",
	"Neon", "NeonClient", "NeonHack", "Koks", "KoksClient",
	"Maf", "MafClient", "MafHack", "PWN", "PWNClient", "PWNHack",
	"PrestigeClient", "Sensation", "SensationClient", "Passion", "PassionClient",
	"Eject", "EjectClient", "EjectHack", "Coffee", "CoffeeClient", "CoffeeHack",
	"MahMeoul", "Meoul", "Insanity", "InsanityClient",
	"Ignite", "IgniteClient", "IgniteHack", "Define", "DefineClient",
	"Nightfall", "NightfallClient", "Vanity", "VanityClient", "Cherish", "CherishClient",
	"Cobalt", "CobaltClient", "CobaltHack", "Equinox", "EquinoxClient",
	// Russian cheat clients
	"Zapret", "ZapretClient", "Raketa", "RaketaClient", "RaketaHack",
	"Pchelin", "PchelinClient", "Sosiska", "SosiskaClient", "SosiskaHack",
	"Shlepa", "ShlepaClient", "ShlepaHack", "Bublik", "BublikClient",
	"GovnoClient", "GovnoHack", "Pivo", "PivoClient",
	"Vodka", "VodkaClient", "CCCP", "CCCPClient", "USSR", "USSRClient",
	"GRU", "GRUClient", "FSB", "FSBClient", "KGB", "KGBClient",
	"ZOV", "ZOVClient", "ZOVHack", "SVO", "SVOClient",
	// More known clients
	"Vega", "VegaClient", "VegaHack", "Toxic", "ToxicClient", "ToxicHack",
	"RageClient", "RageHack", "HackClient", "CheatClient", "ClientHack",
	"R3D", "R3DClient", "R3DHack", "Zanar", "ZanarClient",
	"Pandora", "PandoraClient", "PandoraHack", "Krypton", "KryptonClient",
	"Atlas", "AtlasClient", "AtlasHack", "Origin", "OriginClient",
	"Envy", "EnvyClient", "EnvyHack", "Reaper", "ReaperClient", "ReaperHack",
	"Spigot", "SpigotClient", "SpigotHack", "Melon", "MelonClient",
	"Exter", "ExterClient", "ExterHack", "Mint", "MintClient", "MintHack",
	"Purpur", "PurpurClient", "Endless", "EndlessClient",
	"Dortware", "DortwareClient", "MioHack", "Mio",
	"Senura", "SenuraClient", "Kose", "KoseClient", "KoseHack",
	"Eblow", "EblowClient", "EblowHack", "Trouble", "TroubleClient",
	"RageQuit", "RageQuitClient", "Focus", "FocusClient",
	"Alpha", "AlphaClient", "AlphaHack", "Omega", "OmegaClient", "OmegaHack",
	"Spray", "SprayClient", "SprayHack", "Nemo", "NemoClient",
	"Loon", "LoonClient", "LoonHack", "LoonProject", "Fox", "FoxClient",
}

var broadIndicators = []string{
	"404", "kotlin", "skill", "winner",
	"autofish", "autoeat", "autotool", "autoarmour", "autoattack", "autoclicker",
	"chestesp", "storageesp", "playeresp", "nopush", "jesus", "scaffold",
	"waterclient", "darkproject", "darklight", "decision", "hider", "recode", "fatal",
	"blackberry", "luna", "jessica", "jessia", "excellent", "paragon", "rockstar", "zeus",
	"xray", "dll", "exe", "config",
}

var weakIndicators = makeSet([]string{
	"impact", "future", "fly", "flight", "rise", "delta", "freecam",
	"xray", "x-ray", "esp", "hitbox", "client", "loader", "cheat",
	"meow", "cortex", "lunar", "feather", "badlion", "tlauncher",
	"prismlauncher", "elyprismlauncher", "entropy", "toolbox",
	"regscanner", "usbdeview", "recentfilesview",
	"cherry", "diamond", "spider", "trident", "potions", "moon",
	"invisible", "free", "energy", "destroy", "rich", "wild",
	"simply", "external", "wave", "thunder", "jelly", "infinity",
	"sprint", "neat", "abyss", "bleach", "norender", "autototem",
	"offhand", "killaura", "triggerbot", "aimbot", "nametags",
	"autofish", "autoeat", "autotool", "autoarmour", "autoattack", "autoclicker",
	"chestesp", "storageesp", "playeresp", "nopush", "jesus", "scaffold",
})

var highIndicators = makeSet([]string{
	"vape", "minced", "catlavan", "nursultan", "wexside", "akrien",
	"deadcode", "expensive", "celestial", "venus", "raven", "drip",
	"whiteout", "novoline", "astolfo", "tenacity", "zeroday",
	"exloader", "xenos", "ambientinjector", "liquidloader", "bootject",
	"rename_me_please.dll", "editme.dll", "mc100.dll",
	".vape", ".wex", ".akr", ".bush",
	"vape.gg", "deadcodehack.org", "nursultan.fun",
	"expensiveclient.xyz", "dreampoolhack.ru", "drip.gg",
	"nuclear", "nuclear",
})

var ignoredIndicators = makeSet([]string{
	"labymod", "lunar", "regscanner", "tlauncher",
	"kotlin", "lava", "metro", "elyprismlauncher", "ocean", "neat", "luna",
	"exe", "dll", "config", "404", "skill", "winner",
})

var riskyExt = makeSet([]string{".exe", ".jar", ".dll", ".zip", ".rar", ".7z"})
var textExt = makeSet([]string{".log", ".txt", ".json", ".cfg", ".toml", ".yml", ".yaml", ".properties", ".ini"})
var scanExt = makeSet([]string{".exe", ".jar", ".dll", ".zip", ".rar", ".7z", ".json", ".cfg", ".toml", ".txt", ".log", ".ini"})

var cheatProcessNames = makeSet([]string{
	"energyclient.exe", "energyclient",
	"catlavan.exe", "catlavan",
	"nursultan.exe", "nursultan",
	"wexside.exe", "wexside",
	"nuclear.exe", "nuclear", "nucl3ar.exe",
	"minced.exe", "minced",
	"akrien.exe", "akrien",
	"deadcode.exe", "deadcode",
	"expensive.exe", "expensive",
	"celestial.exe", "celestial",
	"venus.exe", "venus",
	"haruka.exe", "haruka",
	"britva.exe", "britva",
	"delta.exe", "delta",
	"verist.exe", "verist",
	"fluger.exe", "fluger",
	"rastyclient.exe", "rastyclient",
	"nurik.exe", "nurik",
	"matix.exe", "matix",
	"exloader.exe", "exloader",
	"xenos.exe", "xenos",
	"ambientinjector.exe", "ambientinjector",
	"liquidloader.exe", "liquidloader",
	"bootject.exe", "bootject",
	"injector.exe", "injector",
	"vape.exe", "vape",
	"vape.gg",
	"xray.exe",
	"sing-box.exe",
	"horion.exe", "horion",
	"toolbox.exe", "toolbox",
	"zephyr.exe", "zephyr",
	"cortex.exe",
	"interium.exe", "interium",
	"quickclient.exe",
	"wurst.exe", "wurst",
	"meteor.exe", "meteor",
	"liquidbounce.exe", "liquidbounce",
	"sigma.exe", "sigma",
	"impact.exe", "impact",
	"future.exe", "future",
	"aristois.exe", "aristois",
	"inertia.exe", "inertia",
	"bleachhack.exe", "bleachhack",
	"thunderhack.exe", "thunderhack",
	"konas.exe", "konas",
	"raven.exe", "raven",
	"drip.exe", "drip",
	"entropy.exe", "entropy",
	"whiteout.exe", "whiteout",
	"novoline.exe", "novoline",
	"astolfo.exe", "astolfo",
	"tenacity.exe", "tenacity",
	"exhibition.exe", "exhibition",
	"augustus.exe", "augustus",
	"fdp.exe", "fdpclient",
	"nightx.exe", "nightx",
	"azura.exe", "azura",
	"zeroday.exe", "zeroday",
	"aoba.exe", "aoba",
	"jexclient.exe", "jexclient",
	"catalyst.exe", "catalyst",
	"forgebax.exe", "forgehax",
	"forgewurst.exe", "forgewurst",
	"dreampool.exe", "dreampool",
	"tofficlient.exe", "tofficlient",
	"ponos.exe", "ponos",
	"habibimodanalyzer.exe", "habibimodanalyzer",
	"redlotus.exe", "redlotus",
	"doomsday.exe", "doomsdaydetector", "doomsdayfinder",
	"zamorozka.exe", "zamorozka",
	"neverhook.exe", "neverhook",
	"neverware.exe", "neverware",
	"nightmare.exe", "nightmare",
	"summit.exe", "summit",
	"ferox.exe", "ferox",
	"cherry.exe",
	"wintware.exe", "wintware",
	"norules.exe", "norules",
	"eternity.exe", "eternity",
	"archware.exe", "archware",
	"bebraware.exe", "bebraware",
	"nightproject.exe", "nightproject",
	"xone.exe", "xone",
	"venusfree.exe", "venusfree",
	"fanarme.exe", "fanarme",
	"avalone.exe", "avalone",
	"mercu.exe", "mercu",
	"niobium.exe", "niobium",
	"kion.exe", "kion",
	"boberware.exe", "boberware",
	"bushroot.exe", "bushroot",
	"panicattack.exe", "panickattack",
	"hackbrand.exe", "hackbrand",
	"shibeta.exe", "shitbeta",
	"troxill.exe", "troxill",
	"exosware.exe", "exosware",
	"extendopvp.exe", "extendopvp",
	"harambaritone.exe", "harambaritone",
	"flauncher.exe", "flauncher",
	"clownclient.exe", "clownclient",
	"euphoriaclient.exe", "euphoriaclient",
	"vertclient.exe", "vertclient",
	"wortexclient.exe", "wortexclient",
	"mioclient.exe", "mioclient",
	"vegaline.exe", "vegaline",
	"evaware.exe", "evaware",
	"astoria.exe", "astoria",
	"taker.exe", "taker",
	"chameleon.exe", "chameleon",
	"echo-tool.exe", "echotool",
	"lambda.exe", "lambdahack.exe",
	"huzuni.exe", "huzuniclient.exe",
	"fanta.exe", "fantaclient.exe",
	"moneymod.exe", "moneymodclient.exe",
	"bruslik.exe", "bruslikclient.exe",
	"lordcheat.exe", "lordcheatclient.exe",
	"hapu.exe", "hapuclient.exe",
	"champ.exe", "champclient.exe",
	"nfc.exe", "nfcclient.exe",
	"hydra.exe", "hydraclient.exe",
	"phantom.exe", "phantomclient.exe",
	"spooky.exe", "spookyclient.exe",
	"jello.exe", "jelloclient.exe",
	"diablo.exe", "diabloclient.exe",
	"prestige.exe", "prestigeclient.exe",
	"ryu.exe", "ryuclient.exe",
	"prohacker.exe", "prohackerclient.exe",
	"soar.exe", "soarclient.exe",
	"tokaido.exe", "tokaidoclient.exe",
	"rede.exe", "redesky.exe",
	"blizzard.exe", "blizzardclient.exe",
	"strex86.exe", "strex.exe",
	"kuriyama.exe", "kuriyamaclient.exe",
	"sendhook.exe", "sendhookclient.exe",
	"ultraclient.exe", "megaclient.exe", "maxclient.exe",
	"widescreen.exe", "widescreenclient.exe",
	"atani.exe", "ataniclient.exe",
	"yolanda.exe", "yolandaclient.exe",
	"cursed.exe", "cursedclient.exe",
	"minehack.exe", "minehackclient.exe",
	"neon.exe", "neonclient.exe",
	"koks.exe", "koksclient.exe",
	"maf.exe", "mafclient.exe",
	"pwn.exe", "pwnclient.exe",
	"sensation.exe", "sensationclient.exe",
	"passion.exe", "passionclient.exe",
	"eject.exe", "ejectclient.exe",
	"coffee.exe", "coffeeclient.exe",
	"mahmeoul.exe", "meoul.exe",
	"insanity.exe", "insanityclient.exe",
	"ignite.exe", "igniteclient.exe",
	"define.exe", "defineclient.exe",
	"nightfall.exe", "nightfallclient.exe",
	"vanity.exe", "vanityclient.exe",
	"cherish.exe", "cherishclient.exe",
	"cobalt.exe", "cobaltclient.exe",
	"equinox.exe", "equinoxclient.exe",
	"zapret.exe", "zapretclient.exe",
	"raketa.exe", "raketaclient.exe",
	"shlepa.exe", "shlepa.exe", "shlepaclient.exe",
	"bublik.exe", "bublikclient.exe",
	"govnoclient.exe",
	"vega.exe", "vegaclient.exe",
	"toxic.exe", "toxicclient.exe",
	"rageclient.exe", "ragehack.exe",
	"hackclient.exe", "cheatclient.exe",
	"zanar.exe", "zanarclient.exe",
	"pandora.exe", "pandoraclient.exe",
	"krypton.exe", "kryptonclient.exe",
	"atlas.exe", "atlasclient.exe",
	"origin.exe", "originclient.exe",
	"envy.exe", "envyclient.exe",
	"reaper.exe", "reaperclient.exe",
	"exter.exe", "exterclient.exe",
	"mint.exe", "mintclient.exe",
	"purpur.exe", "purpurclient.exe",
	"endless.exe", "endlessclient.exe",
	"dortware.exe", "dortwareclient.exe",
	"senura.exe", "senuraclient.exe",
	"kose.exe", "koseclient.exe",
	"alpha.exe", "alphaclient.exe",
	"omega.exe", "omegaclient.exe",
	"spray.exe", "sprayclient.exe",
	"fox.exe", "foxclient.exe",
})

var cheatDataDirs = []string{
	".vape", ".wex", ".akr", ".bush", ".meteor", ".future",
	"Vape", "Wexside", "Nursultan", "EnergyClient", "Catlavan",
	"NuClear", "Minced", "DeadCode", "Expensive",
}

var minecraftSafeDirs = makeSet([]string{
	"resourcepacks", "dynamic-data-pack-cache", "dynamic-resource-pack-cache",
	"saves", "assets", "libraries", "versions",
})

func main() {
	cfg := parseFlags()
	cutoff := time.Now().AddDate(0, 0, -cfg.days)
	terms := unique(seedIndicators)
	if cfg.broad {
		terms = unique(append(terms, broadIndicators...))
	}
	sort.Slice(terms, func(i, j int) bool { return len(terms[i]) > len(terms[j]) })

	fmt.Println("Go checker scan started...")
	fmt.Printf("Mode: %s | Days: %d | Max files: %d\n", modeName(cfg), cfg.days, cfg.maxFiles)

	findings := runScans(cfg, terms, cutoff)
	data := reportData{GeneratedAt: time.Now(), ComputerName: hostname(), Mode: modeName(cfg), Days: cfg.days, IndicatorsLoaded: len(terms), Findings: findings}
	summary, report := makeReports(data, cutoff)

	stamp := time.Now().Format("20060102_150405")
	outDir := cfg.outDir
	if outDir == "" {
		outDir = os.TempDir()
	}
	_ = os.MkdirAll(outDir, 0755)
	summaryPath := filepath.Join(outDir, "scan_summary_go_"+stamp+".txt")
	reportPath := filepath.Join(outDir, "scan_report_go_"+stamp+".txt")
	jsonPath := filepath.Join(outDir, "scan_report_go_"+stamp+".json")
	mustWrite(summaryPath, []byte(strings.Join(summary, "\r\n")))
	mustWrite(reportPath, []byte(strings.Join(report, "\r\n")))
	if cfg.jsonOut {
		b, _ := json.MarshalIndent(data, "", "  ")
		mustWrite(jsonPath, b)
	}

	fmt.Println("Summary:", summaryPath)
	fmt.Println("Full report:", reportPath)
	if cfg.jsonOut {
		fmt.Println("JSON report:", jsonPath)
	}

	if cfg.telegram {
		token := envDefault("TELEGRAM_BOT_TOKEN", fallbackBotToken)
		chatID := envDefault("TELEGRAM_CHAT_ID", fallbackChatID)
		if err := sendTelegramText(token, chatID, strings.Join(summary, "\n")); err != nil {
			fmt.Println("Telegram summary failed:", err)
		}
		if err := sendTelegramFile(token, chatID, reportPath); err != nil {
			fmt.Println("Telegram report failed:", err)
		} else {
			fmt.Println("Telegram: attempted")
		}
	}
	if cfg.selfDel {
		scheduleSelfDelete()
	}
}

func parseFlags() config {
	var cfg config
	flag.BoolVar(&cfg.telegram, "telegram", true, "send summary and report to Telegram")
	flag.BoolVar(&cfg.broad, "broad", true, "enable wider, noisier indicator list")
	flag.BoolVar(&cfg.deep, "deep", true, "scan more text files and larger logs")
	flag.BoolVar(&cfg.jsonOut, "json", true, "write JSON report next to TXT report")
	flag.BoolVar(&cfg.selfDel, "self-delete", false, "delete this executable after scan exits")
	flag.StringVar(&cfg.outDir, "out", "", "output directory; default is TEMP")
	flag.IntVar(&cfg.days, "days", 60, "date window for dated artifacts")
	flag.IntVar(&cfg.maxFiles, "max-files", 70000, "maximum files to enumerate")
	flag.IntVar(&cfg.maxMs, "max-ms", 180000, "soft scan timeout in milliseconds")
	flag.Parse()
	if cfg.days <= 0 {
		cfg.days = 60
	}
	if cfg.maxFiles < 1000 {
		cfg.maxFiles = 1000
	}
	return cfg
}

func runScans(cfg config, terms []string, cutoff time.Time) []finding {
	started := time.Now()
	type scanResult struct {
		name string
		rows []finding
	}
	jobs := []struct {
		name string
		fn   func() []finding
	}{
		{"Services", scanServices},
		{"Processes", func() []finding { return scanProcesses(terms) }},
		{"Startup", func() []finding { return scanStartupArtifacts(terms) }},
		{"Scheduled tasks", scanScheduledTasks},
		{"DNS cache", func() []finding { return scanCommandText("DNS cache", 12*time.Second, terms, true, "ipconfig.exe", "/displaydns") }},
		{"SRUDB/Data Usage", func() []finding { return scanDataUsage(terms) }},
		{"Current focused files", func() []finding { return scanCurrentFiles(cfg, terms, cutoff, started) }},
		{"Minecraft logs/configs", func() []finding { return scanMinecraftText(cfg, terms, cutoff, started) }},
		{"TEMP suspicious files", func() []finding { return scanTempFiles(terms, cutoff) }},
		{"Browser artifacts", func() []finding { return scanBrowserArtifacts(cfg, terms, cutoff) }},
		{"Recycle Bin", func() []finding { return scanRecycleBin(terms, cutoff) }},
		{"Recent", func() []finding { return scanSimpleDir("Recent", filepath.Join(env("APPDATA"), "Microsoft", "Windows", "Recent"), terms, cutoff) }},
		{"Prefetch", func() []finding { return scanSimpleDir("Prefetch", filepath.Join(envDefault("SystemRoot", `C:\Windows`), "Prefetch"), terms, cutoff) }},
		{"Manual files", func() []finding { return scanManualFiles(terms) }},
		{"Cheat data dirs", func() []finding { return scanCheatDataDirs(terms, cutoff) }},
		{"Installed programs", func() []finding { return scanInstalledPrograms(terms) }},
	}

	results := make(chan scanResult, len(jobs))
	var wg sync.WaitGroup
	sem := make(chan struct{}, 4)
	for _, job := range jobs {
		wg.Add(1)
		go func(name string, fn func() []finding) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			fmt.Println("Scanning", name+"...")
			results <- scanResult{name: name, rows: fn()}
		}(job.name, job.fn)
	}
	wg.Wait()
	close(results)

	all := []finding{}
	for res := range results {
		all = append(all, res.rows...)
	}
	sortFindings(all)
	return dedupeFindings(all)
}

func scanServices() []finding {
	names := []string{"PcaSvc", "CDPSvc", "DPS", "SSDPSRV", "DiagTrack", "SysMain", "EventLog", "Appinfo", "WSearch", "DusmSvc"}
	out := []finding{}
	for _, name := range names {
		b, _ := commandOutput(6*time.Second, "sc.exe", "query", name)
		state := "UNKNOWN"
		s := strings.ToUpper(string(b))
		if strings.Contains(s, "RUNNING") {
			state = "RUNNING"
		} else if strings.Contains(s, "STOPPED") {
			state = "STOPPED"
		}
		risk := "info"
		if state != "RUNNING" {
			risk = "low"
		}
		out = append(out, finding{Source: "Services", Risk: risk, Matches: []string{name}, Details: "State: " + state})
	}
	return out
}

func scanProcesses(terms []string) []finding {
	b, err := commandOutput(20*time.Second, "wmic", "process", "get", "ProcessId,Name,ExecutablePath,CommandLine", "/format:csv")
	if err != nil && len(b) == 0 {
		b, _ = commandOutput(20*time.Second, "powershell.exe", "-NoProfile", "-Command", "Get-CimInstance Win32_Process | Select ProcessId,Name,ExecutablePath,CommandLine | Format-List")
	}
	out := []finding{}
	seen := map[string]bool{}
	for _, block := range splitCommandRows(string(b)) {
		processName := extractProcessName(block)
		// Check 1: process name is a known cheat process
		if cheatProcessNames[strings.ToLower(processName)] {
			nameMatches := []string{processName}
			out = append(out, finding{Source: "Processes", Risk: "high", Matches: nameMatches, Details: trim(block, 1200)})
			continue
		}
		// Check 2: command line or full block contains cheat indicators
		m := matchText(block, terms, false)
		if len(m) > 0 {
			key := strings.ToLower(processName)
			if !seen[key] {
				seen[key] = true
				out = append(out, finding{Source: "Processes", Risk: riskOf(m), Matches: m, Details: trim(block, 1200)})
			}
		}
	}
	return limitFindings(out, 150)
}

func extractProcessName(block string) string {
	lines := strings.Split(block, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(strings.ToLower(line), "name") || strings.HasPrefix(strings.ToLower(line), "name=") {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) < 2 {
				parts = strings.SplitN(line, ":", 2)
			}
			if len(parts) == 2 {
				return strings.TrimSpace(parts[1])
			}
		}
	}
	if strings.Contains(block, ",") {
		fields := strings.Split(block, ",")
		if len(fields) >= 2 {
			return strings.Trim(fields[1], "\" ")
		}
	}
	return ""
}

func scanStartupArtifacts(terms []string) []finding {
	commands := [][]string{
		{"reg.exe", "query", `HKCU\Software\Microsoft\Windows\CurrentVersion\Run`, "/s"},
		{"reg.exe", "query", `HKCU\Software\Microsoft\Windows\CurrentVersion\RunOnce`, "/s"},
		{"reg.exe", "query", `HKLM\Software\Microsoft\Windows\CurrentVersion\Run`, "/s"},
		{"reg.exe", "query", `HKLM\Software\Microsoft\Windows\CurrentVersion\RunOnce`, "/s"},
		{"reg.exe", "query", `HKLM\Software\WOW6432Node\Microsoft\Windows\CurrentVersion\Run`, "/s"},
	}
	out := []finding{}
	for _, cmd := range commands {
		b, _ := commandOutput(8*time.Second, cmd[0], cmd[1:]...)
		for _, row := range splitCommandRows(string(b)) {
			m := matchText(row, terms, false)
			if len(m) > 0 {
				out = append(out, finding{Source: "Startup", Risk: riskOf(m), Matches: m, Details: trim(row, 1000)})
			}
		}
	}
	startupDirs := []string{filepath.Join(env("APPDATA"), "Microsoft", "Windows", "Start Menu", "Programs", "Startup"), filepath.Join(env("ProgramData"), "Microsoft", "Windows", "Start Menu", "Programs", "Startup")}
	for _, row := range scanPaths("Startup", startupDirs, terms, nil, 2000, time.Time{}) {
		out = append(out, row)
	}
	startupTiming := scanProcessTiming()
	out = append(out, startupTiming...)
	return limitFindings(out, 120)
}

func scanProcessTiming() []finding {
	methods := []struct {
		name string
		cmd  []string
	}{
		{"wmic", []string{"wmic", "process", "get", "Name,ProcessId,CreationDate", "/format:csv"}},
		{"ps_csv", []string{"powershell.exe", "-NoProfile", "-Command", "Get-CimInstance Win32_Process | Select-Object Name,ProcessId,CreationDate | ConvertTo-Csv -NoTypeInformation 2>$null"}},
		{"ps_flist", []string{"powershell.exe", "-NoProfile", "-Command", "Get-CimInstance Win32_Process | Where-Object { $_.Name -eq 'explorer.exe' -or $_.Name -eq 'javaw.exe' -or $_.Name -eq 'java.exe' } | Format-List Name,ProcessId,CreationDate"}},
	}
	var b []byte
	for _, m := range methods {
		b2, err2 := commandOutput(20*time.Second, m.cmd[0], m.cmd[1:]...)
		if len(b2) > 0 {
			b = b2
			if err2 == nil {
				break
			}
		}
	}
	if len(b) == 0 {
		return nil
	}
	out := []finding{}
	for _, row := range strings.Split(strings.ReplaceAll(string(b), "\r\n", "\n"), "\n") {
		row = strings.TrimSpace(row)
		if row == "" {
			continue
		}
		if strings.HasPrefix(row, "\"Name\"") || strings.HasPrefix(row, "Name") || strings.HasPrefix(row, "#TYPE") || strings.HasPrefix(row, "\"CLASS\"") {
			continue
		}
		if strings.HasPrefix(strings.ToLower(row), "node") {
			continue
		}
		var name, creationDate string
		if strings.Contains(row, ",") {
			fields := splitCSV(row)
			if len(fields) >= 3 && !isAllDigits(fields[0]) {
				name = fields[0]
				creationDate = fields[len(fields)-1]
			} else if len(fields) >= 3 {
				name = fields[1]
				creationDate = fields[len(fields)-1]
			}
		} else if strings.Contains(row, ":") {
			lines := strings.Split(row, "\n")
			for _, line := range lines {
				line = strings.TrimSpace(line)
				parts := strings.SplitN(line, ":", 2)
				if len(parts) != 2 {
					continue
				}
				key := strings.ToLower(strings.TrimSpace(parts[0]))
				val := strings.TrimSpace(parts[1])
				switch key {
				case "name":
					name = val
				case "creationdate":
					creationDate = val
				}
			}
		}
		if name == "" || creationDate == "" || len(creationDate) < 14 {
			continue
		}
		lowName := strings.ToLower(name)
		if !strings.Contains(lowName, "explorer") && !strings.Contains(lowName, "java") {
			continue
		}
		t, err := time.Parse("20060102150405", creationDate[:14])
		if err != nil {
			continue
		}
		out = append(out, finding{
			Source:  "Startup",
			Risk:    "info",
			Matches: []string{name},
			Details: fmt.Sprintf("%s - %s", name, t.Format("2006.01.02 15:04:05")),
		})
	}
	return out
}

func isAllDigits(s string) bool {
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return len(s) > 0
}

func scanScheduledTasks() []finding {
	b, err := commandOutput(25*time.Second, "powershell.exe", "-NoProfile", "-Command", "Get-ScheduledTask | Where-Object { $_.State -ne 'Disabled' } | Select-Object TaskName,TaskPath,State | ConvertTo-Csv -NoTypeInformation")
	if err != nil || len(b) == 0 {
		return []finding{{Source: "Scheduled tasks", Risk: "info", Details: "No scheduled tasks data available"}}
	}
	out := []finding{}
	rows := strings.Split(strings.ReplaceAll(string(b), "\r\n", "\n"), "\n")
	for _, row := range rows {
		row = strings.TrimSpace(row)
		if row == "" || strings.HasPrefix(row, "TaskName") || strings.HasPrefix(row, "\"TaskName\"") {
			continue
		}
		fields := splitCSV(row)
		if len(fields) < 3 {
			continue
		}
		taskName := fields[0]
		taskPath := fields[1]
		status := fields[2]
		if strings.EqualFold(status, "Disabled") {
			continue
		}
		taskPathLower := strings.ToLower(taskPath)
		if strings.HasPrefix(taskPathLower, `\microsoft\`) ||
			strings.HasPrefix(taskPathLower, `\google`) ||
			strings.HasPrefix(taskPathLower, `\nvidia`) ||
			strings.HasPrefix(taskPathLower, `\bluestacks`) ||
			strings.HasPrefix(taskPathLower, `\adobe`) ||
			strings.Contains(taskPathLower, `\windows\`) {
			continue
		}
		detail := fmt.Sprintf("TaskName: %s | TaskPath: %s | State: %s", taskName, taskPath, status)
		out = append(out, finding{Source: "Scheduled tasks", Risk: "info", Details: detail})
	}
	if len(out) == 0 {
		return []finding{{Source: "Scheduled tasks", Risk: "info", Details: "No non-system scheduled tasks found"}}
	}
	return limitFindings(out, 100)
}

func splitCSV(line string) []string {
	var fields []string
	inQuote := false
	current := ""
	for _, ch := range line {
		if ch == '"' {
			inQuote = !inQuote
			continue
		}
		if ch == ',' && !inQuote {
			fields = append(fields, current)
			current = ""
			continue
		}
		current += string(ch)
	}
	fields = append(fields, current)
	return fields
}

func scanCommandText(source string, timeout time.Duration, terms []string, contentMode bool, name string, args ...string) []finding {
	b, err := commandOutput(timeout, name, args...)
	if err != nil && len(b) == 0 {
		return []finding{{Source: source, Risk: "info", Details: "command unavailable or timed out"}}
	}
	out := []finding{}
	for _, row := range splitCommandRows(string(b)) {
		m := matchText(row, terms, contentMode)
		if len(m) > 0 {
			out = append(out, finding{Source: source, Risk: riskOf(m), Matches: m, Details: trim(row, 1200)})
		}
	}
	return limitFindings(out, 150)
}

func scanCurrentFiles(cfg config, terms []string, cutoff time.Time, started time.Time) []finding {
	home, _ := os.UserHomeDir()
	roots := uniquePaths([]string{
		filepath.Join(home, "Desktop"), filepath.Join(home, "Downloads"), filepath.Join(home, "Documents"),
		env("APPDATA"), env("LOCALAPPDATA"), env("ProgramData"), env("ProgramFiles"), env("ProgramFiles(x86)"),
	})
	return scanPaths("Current focused files", roots, terms, scanExt, cfg.maxFiles, cutoff, started, time.Duration(cfg.maxMs)*time.Millisecond, 500)
}

func scanMinecraftText(cfg config, terms []string, cutoff time.Time, started time.Time) []finding {
	roots := uniquePaths([]string{
		filepath.Join(env("APPDATA"), ".minecraft"), filepath.Join(env("APPDATA"), ".tlauncher"), filepath.Join(env("APPDATA"), ".lunarclient"), filepath.Join(env("APPDATA"), ".feather"), filepath.Join(env("APPDATA"), "ElyPrismLauncher"),
	})
	maxSize := int64(2 * 1024 * 1024)
	maxRows := 200
	if cfg.deep {
		maxSize = 8 * 1024 * 1024
		maxRows = 500
	}
	files := walkRoots(roots, cfg.maxFiles/2, func(p string, d os.DirEntry) bool { return !d.IsDir() && textExt[strings.ToLower(filepath.Ext(p))] }, started, time.Duration(cfg.maxMs)*time.Millisecond)
	out := []finding{}
	for _, p := range files {
		if isKnownSafePath(p) || isMinecraftAssetIndex(p) || isMinecraftFPPath(p) {
			continue
		}
		st, err := os.Stat(p)
		if err != nil || st.ModTime().Before(cutoff) || st.Size() > maxSize {
			continue
		}
		b, err := readLimited(p, maxSize)
		if err != nil {
			continue
		}
		m := matchText(p+"\n"+string(b), terms, true)
		if len(m) > 0 {
			filtered := filterMinecraftMatches(m)
			if len(filtered) > 0 {
				out = append(out, finding{Source: "Minecraft logs/configs", Risk: riskOf(filtered), Matches: filtered, Path: p, Size: st.Size(), Modified: st.ModTime()})
			}
		}
		if len(out) >= maxRows {
			break
		}
	}
	return out
}

func isMinecraftFPPath(p string) bool {
	pl := strings.ToLower(filepath.Clean(p))
	fpParts := []string{
		`\resourcepacks\`, `\dynamic-data-pack-cache\`, `\dynamic-resource-pack-cache\`,
		`\saves\`, `\assets\indexes\`,
		`\libraries\`, `\versions\`, `\minecraft\bin\`,
		`\configureddefaults\`, `\flame\overrides.txt`, `\mrpack\overrides.txt`,
	}
	for _, part := range fpParts {
		if strings.Contains(pl, part) {
			return true
		}
	}
	return false
}

var minecraftSafeContent = makeSet([]string{
	"cherry", "diamond", "spider", "trident", "potions", "moon",
	"invisible", "free", "energy", "destroy", "rich", "wild",
	"simply", "external", "wave", "thunder", "jelly", "infinity",
	"sprint", "neat", "abyss", "bleach",
})

func filterMinecraftMatches(matches []string) []string {
	out := []string{}
	for _, m := range matches {
		if !minecraftSafeContent[strings.ToLower(m)] {
			out = append(out, m)
		}
	}
	return out
}

func scanTempFiles(terms []string, cutoff time.Time) []finding {
	roots := uniquePaths([]string{os.TempDir(), env("TEMP"), env("TMP"), filepath.Join(env("LOCALAPPDATA"), "Temp")})
	return scanPaths("TEMP suspicious files", roots, terms, riskyExt, 12000, cutoff, time.Now(), 45*time.Second, 200)
}

func scanRecycleBin(terms []string, cutoff time.Time) []finding {
	roots := uniquePaths([]string{filepath.Join(envDefault("SystemDrive", `C:`), `$Recycle.Bin`)})
	return scanPaths("Recycle Bin", roots, terms, scanExt, 8000, cutoff, time.Now(), 25*time.Second, 100)
}

func scanDataUsage(terms []string) []finding {
	src := filepath.Join(envDefault("SystemRoot", `C:\Windows`), "System32", "sru", "SRUDB.dat")
	st, err := os.Stat(src)
	if err != nil {
		return []finding{{Source: "SRUDB/Data Usage", Risk: "info", Details: "SRUDB.dat unavailable: " + err.Error()}}
	}
	tmp := filepath.Join(os.TempDir(), fmt.Sprintf("SRUDB_checker_%d.dat", time.Now().UnixNano()))
	if err := copyFile(src, tmp); err != nil {
		return []finding{{Source: "SRUDB/Data Usage", Risk: "info", Details: "SRUDB.dat copy failed: " + err.Error()}}
	}
	defer os.Remove(tmp)
	b, err := readLimited(tmp, 160*1024*1024)
	if err != nil {
		return []finding{{Source: "SRUDB/Data Usage", Risk: "info", Details: "SRUDB.dat read failed: " + err.Error()}}
	}
	m := matchText(string(b), terms, true)
	if len(m) == 0 {
		return nil
	}
	return []finding{{Source: "SRUDB/Data Usage", Risk: riskOf(m), Matches: m, Path: src, Size: st.Size(), Modified: st.ModTime(), Details: "binary string hit in Windows Data Usage database; date inside SRUDB is not parsed by Go scanner"}}
}

func scanBrowserArtifacts(cfg config, terms []string, cutoff time.Time) []finding {
	browserTerms := browserIndicatorTerms(terms)
	roots := uniquePaths([]string{
		filepath.Join(env("LOCALAPPDATA"), "Google", "Chrome", "User Data"),
		filepath.Join(env("LOCALAPPDATA"), "Microsoft", "Edge", "User Data"),
		filepath.Join(env("APPDATA"), "Mozilla", "Firefox", "Profiles"),
	})
	names := makeSet([]string{"history", "login data", "cookies", "top sites", "places.sqlite", "formhistory.sqlite"})
	files := walkRoots(roots, 1800, func(p string, d os.DirEntry) bool { return !d.IsDir() && names[strings.ToLower(filepath.Base(p))] }, time.Now(), 45*time.Second)
	out := []finding{}
	for _, p := range files {
		st, err := os.Stat(p)
		if err != nil || st.ModTime().Before(cutoff) || st.Size() > 90*1024*1024 {
			continue
		}
		b, err := readLimited(p, 90*1024*1024)
		if err != nil {
			continue
		}
		m := matchText(string(b), browserTerms, false)
		if len(m) > 0 {
			out = append(out, finding{Source: "Browser artifacts", Risk: riskOf(m), Matches: m, Path: p, Size: st.Size(), Modified: st.ModTime()})
		}
		if len(out) >= 80 {
			break
		}
	}
	return out
}

func scanSimpleDir(source, root string, terms []string, cutoff time.Time) []finding {
	return scanPaths(source, []string{root}, terms, nil, 6000, cutoff, time.Now(), 25*time.Second, 120)
}

func scanCheatDataDirs(terms []string, cutoff time.Time) []finding {
	home, _ := os.UserHomeDir()
	appdata := env("APPDATA")
	localAppdata := env("LOCALAPPDATA")
	roots := []string{}
	for _, dir := range cheatDataDirs {
		paths := []string{
			filepath.Join(home, dir),
			filepath.Join(appdata, dir),
			filepath.Join(appdata, ".minecraft", dir),
			filepath.Join(localAppdata, dir),
			filepath.Join(os.TempDir(), dir),
		}
		roots = append(roots, paths...)
	}
	return scanPaths("Cheat data dirs", roots, terms, nil, 5000, cutoff, time.Now(), 30*time.Second, 100)
}

func scanInstalledPrograms(terms []string) []finding {
	regPaths := []string{
		`HKLM\Software\Microsoft\Windows\CurrentVersion\Uninstall`,
		`HKLM\Software\WOW6432Node\Microsoft\Windows\CurrentVersion\Uninstall`,
		`HKCU\Software\Microsoft\Windows\CurrentVersion\Uninstall`,
	}
	out := []finding{}
	for _, regPath := range regPaths {
		b, err := commandOutput(10*time.Second, "reg.exe", "query", regPath, "/s", "/f", "", "/e")
		if err != nil || len(b) == 0 {
			continue
		}
		for _, row := range splitCommandRows(string(b)) {
			m := matchText(row, terms, false)
			if len(m) > 0 {
				out = append(out, finding{Source: "Installed programs", Risk: riskOf(m), Matches: m, Details: trim(row, 500)})
			}
		}
		if len(out) >= 50 {
			break
		}
	}
	return out
}

func scanManualFiles(terms []string) []finding {
	home, _ := os.UserHomeDir()
	desktop := filepath.Join(home, "Desktop")
	names := []string{"Manual.txt", "manual.txt", "cheats.txt", "\u041c\u0430\u043d\u0443\u0430\u043b.txt", "\u043c\u0430\u043d\u0443\u0430\u043b 2.txt", "\u043c\u0430\u043d\u0443\u0430\u043b 3.txt", "\u0447\u0438\u0442\u044b.txt"}
	out := []finding{}
	for _, name := range names {
		p := filepath.Join(desktop, name)
		st, err := os.Stat(p)
		if err != nil || st.Size() > 10*1024*1024 {
			continue
		}
		b, err := readLimited(p, 10*1024*1024)
		if err != nil {
			continue
		}
		m := matchText(string(b), terms, false)
		if len(m) > 0 {
			out = append(out, finding{Source: "Manual files", Risk: "info", Path: p, Size: st.Size(), Modified: st.ModTime(), Details: fmt.Sprintf("manual indicator file contains %d known terms; used as context, not evidence", len(m))})
		}
	}
	return out
}

func scanPaths(source string, roots []string, terms []string, exts map[string]bool, maxFiles int, cutoff time.Time, args ...interface{}) []finding {
	started := time.Now()
	timeout := 60 * time.Second
	limit := 250
	if len(args) > 0 {
		if t, ok := args[0].(time.Time); ok && !t.IsZero() {
			started = t
		}
	}
	if len(args) > 1 {
		if d, ok := args[1].(time.Duration); ok && d > 0 {
			timeout = d
		}
	}
	if len(args) > 2 {
		if n, ok := args[2].(int); ok && n > 0 {
			limit = n
		}
	}
	files := walkRoots(roots, maxFiles, func(p string, d os.DirEntry) bool {
		if d.IsDir() {
			return false
		}
		if exts != nil && !exts[strings.ToLower(filepath.Ext(p))] {
			return false
		}
		return true
	}, started, timeout)
	out := []finding{}
	seen := map[string]bool{}
	for _, p := range files {
		if seen[p] {
			continue
		}
		seen[p] = true
		st, err := os.Stat(p)
		if err != nil {
			continue
		}
		if !cutoff.IsZero() && st.ModTime().Before(cutoff) {
			continue
		}
		m := matchPath(p, terms)
		if len(m) == 0 {
			continue
		}
		detail := "recent file"
		if !cutoff.IsZero() && st.ModTime().Before(cutoff) {
			detail = "current file, old timestamp"
		}
		out = append(out, finding{Source: source, Risk: riskOf(m), Matches: m, Path: p, Size: st.Size(), Modified: st.ModTime(), Details: detail})
		if len(out) >= limit {
			break
		}
	}
	return out
}

func walkRoots(roots []string, limit int, accept func(string, os.DirEntry) bool, started time.Time, timeout time.Duration) []string {
	skipParts := []string{
		`\node_modules`, `\.git`, `\Cache`, `\GPUCache`, `\WindowsApps`,
		`\Service Worker`, `\Code\Cache`, `\Python\pythoncore-`, `\Temp\scoped_dir`,
		`\AppData\Local\Packages`,
		`\EdgeCore`, `\EdgeUpdate`, `\EdgeWebView`,
		`\Microsoft\Edge\Application`, `\Microsoft\EdgeCore`, `\Microsoft\EdgeUpdate`,
		`\Microsoft\EdgeWebView`,
		`\Reference Assemblies`,
		`\Windows Defender\Definition Updates`,
		`\Windows Defender\Platform`,
		`\Microsoft\Windows\OneSettings`,
	}
	out := []string{}
	var mu sync.Mutex
	var wg sync.WaitGroup
	sem := make(chan struct{}, 4)
	for _, root := range roots {
		if root == "" {
			continue
		}
		if _, err := os.Stat(root); err != nil {
			continue
		}
		wg.Add(1)
		go func(root string) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			_ = filepath.WalkDir(root, func(p string, d os.DirEntry, err error) error {
				if err != nil {
					return nil
				}
				if timeout > 0 && time.Since(started) > timeout {
					return errors.New("timeout")
				}
				pl := strings.ToLower(p)
				if d.IsDir() {
					for _, s := range skipParts {
						if strings.Contains(pl, strings.ToLower(s)) {
							return filepath.SkipDir
						}
					}
					return nil
				}
				mu.Lock()
				full := len(out) >= limit
				mu.Unlock()
				if full {
					return errors.New("limit")
				}
				if accept(p, d) {
					mu.Lock()
					if len(out) < limit {
						out = append(out, p)
					}
					mu.Unlock()
				}
				return nil
			})
		}(root)
	}
	wg.Wait()
	return out
}

func matchPath(p string, terms []string) []string {
	if isKnownSafePath(p) {
		return nil
	}
	base := strings.ToLower(filepath.Base(p))
	full := strings.ToLower(p)
	baseLoose := looseText(base)
	fullLoose := looseText(full)
	ext := strings.ToLower(filepath.Ext(p))
	found := map[string]string{}
	for _, term := range terms {
		low := strings.ToLower(term)
		loose := looseText(low)
		if ignoredIndicators[low] || (low == "norender" && ext == ".dll") {
			continue
		}
		if strings.HasPrefix(low, ".") {
			if containsBounded(full, low, needsBoundary(term)) {
				found[low] = term
			}
			continue
		}
		if weakIndicators[low] {
			if riskyExt[ext] && (containsBounded(base, low, needsBoundary(term)) || containsBounded(baseLoose, loose, needsBoundary(term))) {
				found[low] = term
			}
			continue
		}
		if containsBounded(full, low, needsBoundary(term)) || containsBounded(fullLoose, loose, needsBoundary(term)) {
			found[low] = term
		}
	}
	return sortedValues(found)
}

func looseText(s string) string {
	replacer := strings.NewReplacer("_", " ", "-", " ", ".", " ", "(", " ", ")", " ", "[", " ", "]", " ")
	return strings.Join(strings.Fields(replacer.Replace(strings.ToLower(s))), " ")
}

func browserIndicatorTerms(terms []string) []string {
	out := []string{}
	for _, term := range terms {
		low := strings.ToLower(term)
		if ignoredIndicators[low] {
			continue
		}
		if strings.Contains(low, ".") && !strings.HasPrefix(low, ".") && !weakIndicators[low] {
			out = append(out, term)
		}
	}
	return unique(out)
}

func isKnownSafePath(p string) bool {
	pl := strings.ToLower(filepath.Clean(p))
	safeParts := []string{
		`\program files\go\src\`, `\program files\atlas toolbox\`,
		`\python\pythoncore-`, `\doc\html\_sources\`, `\nvidia overlay\`,
		`\windows\system32\sru\`,
	}
	for _, part := range safeParts {
		if strings.Contains(pl, strings.ToLower(part)) {
			return true
		}
	}
	return false
}

func isMinecraftAssetIndex(p string) bool {
	pl := strings.ToLower(filepath.Clean(p))
	return strings.Contains(pl, `\assets\indexes\`) && strings.HasSuffix(pl, ".json")
}

func matchText(text string, terms []string, contentMode bool) []string {
	lowText := strings.ToLower(text)
	found := map[string]string{}
	for _, term := range terms {
		low := strings.ToLower(term)
		if ignoredIndicators[low] {
			continue
		}
		if contentMode && weakIndicators[low] {
			continue
		}
		if containsBounded(lowText, low, needsBoundary(term)) {
			found[low] = term
		}
	}
	return sortedValues(found)
}

func containsBounded(text, needle string, bounded bool) bool {
	start := 0
	for {
		i := strings.Index(text[start:], needle)
		if i < 0 {
			return false
		}
		pos := start + i
		if !bounded || (isBoundary(text, pos-1) && isBoundary(text, pos+len(needle))) {
			return true
		}
		start = pos + len(needle)
		if start >= len(text) {
			return false
		}
	}
}

func isBoundary(s string, idx int) bool {
	if idx < 0 || idx >= len(s) {
		return true
	}
	c := s[idx]
	return !((c >= 'a' && c <= 'z') || (c >= '0' && c <= '9'))
}

func needsBoundary(s string) bool {
	if s == "" {
		return false
	}
	first, last := s[0], s[len(s)-1]
	return isAlphaNum(first) && isAlphaNum(last)
}

func isAlphaNum(c byte) bool {
	return (c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9')
}

func riskOf(matches []string) string {
	risk := "medium"
	for _, m := range matches {
		low := strings.ToLower(m)
		if highIndicators[low] {
			return "high"
		}
		if weakIndicators[low] {
			risk = "low"
		}
	}
	return risk
}

func makeReports(data reportData, cutoff time.Time) ([]string, []string) {
	summary := []string{
		fmt.Sprintf("Время проверки: %s", formatDate(data.GeneratedAt)),
		"Движок: Go scanner",
		"ПК: " + data.ComputerName,
		"Режим: " + russianMode(data.Mode),
		fmt.Sprintf("Индикаторов загружено: %d", data.IndicatorsLoaded),
		"Фильтр дат: последние " + fmt.Sprint(data.Days) + " дней, начиная с " + cutoff.Format("2006.01.02"),
	}
	bySource := map[string]map[string]bool{}
	riskCounts := map[string]int{"high": 0, "medium": 0, "low": 0, "info": 0}
	for _, f := range data.Findings {
		riskCounts[f.Risk]++
		if f.Risk == "info" || len(f.Matches) == 0 {
			continue
		}
		if bySource[f.Source] == nil {
			bySource[f.Source] = map[string]bool{}
		}
		for _, m := range f.Matches {
			bySource[f.Source][m] = true
		}
	}
	summary = append(summary, fmt.Sprintf("Уровни риска: высокий=%d средний=%d низкий=%d инфо=%d", riskCounts["high"], riskCounts["medium"], riskCounts["low"], riskCounts["info"]))
	keys := sortedMapKeys(bySource)
	for _, k := range keys {
		summary = append(summary, fmt.Sprintf("%s - найдено: %s", russianSource(k), strings.Join(sortedBoolKeys(bySource[k]), ", ")))
	}
	if len(keys) == 0 {
		summary = append(summary, "Индикаторы читов - ничего")
	}

	report := []string{
		fmt.Sprintf("Go scan report: %s", formatDate(data.GeneratedAt)),
		"PC: " + data.ComputerName,
		"Mode: " + data.Mode,
		fmt.Sprintf("Indicators loaded: %d", data.IndicatorsLoaded),
	}
	groups := map[string][]finding{}
	for _, f := range data.Findings {
		groups[f.Source] = append(groups[f.Source], f)
	}
	sections := []string{"Services", "Processes", "Startup", "Scheduled tasks", "DNS cache", "SRUDB/Data Usage", "Current focused files", "Minecraft logs/configs", "TEMP suspicious files", "Browser artifacts", "Recycle Bin", "Recent", "Prefetch", "Manual files", "Cheat data dirs", "Installed programs"}
	for _, s := range sections {
		report = append(report, "", "==== "+s+" ("+russianSource(s)+") ====")
		rows := groups[s]
		if len(rows) == 0 {
			report = append(report, "No results.")
			continue
		}
		for _, f := range rows {
			report = append(report, formatFinding(f))
		}
	}
	report = append(report, "", "Note: Go scanner v2 - improved accuracy, reduced noise in Minecraft configs and generic paths.")
	return summary, report
}

func formatFinding(f finding) string {
	parts := []string{"risk: " + f.Risk}
	if len(f.Matches) > 0 {
		parts = append(parts, "matches: "+strings.Join(f.Matches, ", "))
	}
	if f.Path != "" {
		parts = append(parts, "path: "+f.Path)
	}
	if f.Size > 0 {
		parts = append(parts, fmt.Sprintf("size: %d", f.Size))
	}
	if !f.Modified.IsZero() {
		parts = append(parts, "modified: "+formatDate(f.Modified))
	}
	if f.Details != "" {
		parts = append(parts, "details: "+f.Details)
	}
	return strings.Join(parts, " | ")
}

func commandOutput(timeout time.Duration, name string, args ...string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.SysProcAttr = nil
	b, err := cmd.CombinedOutput()
	if ctx.Err() == context.DeadlineExceeded {
		return b, ctx.Err()
	}
	return b, err
}

func scheduleSelfDelete() {
	exe, err := os.Executable()
	if err != nil || exe == "" {
		return
	}
	cmdLine := fmt.Sprintf("ping 127.0.0.1 -n 2 >nul & del /f /q \"%s\"", exe)
	_ = exec.Command("cmd.exe", "/C", cmdLine).Start()
}

func splitCommandRows(s string) []string {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	chunks := strings.Split(s, "\n\n")
	if len(chunks) <= 1 {
		chunks = strings.Split(s, "\n")
	}
	out := []string{}
	for _, c := range chunks {
		c = strings.TrimSpace(c)
		if c != "" {
			out = append(out, c)
		}
	}
	return out
}

func readLimited(p string, max int64) ([]byte, error) {
	f, err := os.Open(p)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return io.ReadAll(io.LimitReader(f, max+1))
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	_, copyErr := io.Copy(out, in)
	closeErr := out.Close()
	if copyErr != nil {
		return copyErr
	}
	return closeErr
}

func formatDate(t time.Time) string {
	return t.Format("2006.01.02 15:04:05")
}

func sendTelegramText(token, chatID, text string) error {
	if len(text) > 3900 {
		text = text[:3900] + "\n...trimmed"
	}
	resp, err := http.PostForm("https://api.telegram.org/bot"+token+"/sendMessage", mapValues(map[string]string{"chat_id": chatID, "text": text, "disable_web_page_preview": "true"}))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return fmt.Errorf("http %s", resp.Status)
	}
	return nil
}

func sendTelegramFile(token, chatID, filePath string) error {
	var body bytes.Buffer
	w := multipart.NewWriter(&body)
	_ = w.WriteField("chat_id", chatID)
	_ = w.WriteField("caption", "Full report "+hostname())
	fw, err := w.CreateFormFile("document", filepath.Base(filePath))
	if err != nil {
		return err
	}
	f, err := os.Open(filePath)
	if err != nil {
		return err
	}
	_, err = io.Copy(fw, f)
	_ = f.Close()
	if err != nil {
		return err
	}
	_ = w.Close()
	req, err := http.NewRequest("POST", "https://api.telegram.org/bot"+token+"/sendDocument", &body)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", w.FormDataContentType())
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return fmt.Errorf("http %s", resp.Status)
	}
	return nil
}

func sortFindings(rows []finding) {
	riskRank := map[string]int{"high": 0, "medium": 1, "low": 2, "info": 3}
	sort.Slice(rows, func(i, j int) bool {
		if rows[i].Source != rows[j].Source {
			return rows[i].Source < rows[j].Source
		}
		if riskRank[rows[i].Risk] != riskRank[rows[j].Risk] {
			return riskRank[rows[i].Risk] < riskRank[rows[j].Risk]
		}
		return rows[i].Path < rows[j].Path
	})
}

func dedupeFindings(rows []finding) []finding {
	seen := map[string]bool{}
	out := []finding{}
	for _, r := range rows {
		key := r.Source + "\x00" + r.Risk + "\x00" + strings.Join(r.Matches, ",") + "\x00" + r.Path + "\x00" + r.Details
		if seen[key] {
			continue
		}
		seen[key] = true
		out = append(out, r)
	}
	return out
}

func limitFindings(rows []finding, n int) []finding {
	if len(rows) <= n {
		return rows
	}
	return rows[:n]
}

func mapValues(m map[string]string) url.Values {
	out := url.Values{}
	for k, v := range m {
		out.Set(k, v)
	}
	return out
}

func makeSet(xs []string) map[string]bool {
	m := map[string]bool{}
	for _, x := range xs {
		m[strings.ToLower(x)] = true
	}
	return m
}

func unique(xs []string) []string {
	m := map[string]string{}
	for _, x := range xs {
		if x != "" {
			m[strings.ToLower(x)] = x
		}
	}
	return sortedValues(m)
}

func uniquePaths(xs []string) []string {
	m := map[string]string{}
	for _, x := range xs {
		if x != "" {
			m[strings.ToLower(filepath.Clean(x))] = filepath.Clean(x)
		}
	}
	return sortedValues(m)
}

func sortedValues(m map[string]string) []string {
	out := []string{}
	for _, v := range m {
		out = append(out, v)
	}
	sort.Strings(out)
	return out
}

func sortedBoolKeys(m map[string]bool) []string {
	out := []string{}
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

func sortedMapKeys(m map[string]map[string]bool) []string {
	out := []string{}
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

func env(name string) string { return os.Getenv(name) }

func envDefault(name, fallback string) string {
	if v := os.Getenv(name); v != "" {
		return v
	}
	return fallback
}

func modeName(cfg config) string {
	if cfg.broad {
		return "broad"
	}
	return "focused"
}

func russianMode(mode string) string {
	if mode == "broad" {
		return "расширенный"
	}
	return "точечный"
}

func russianSource(source string) string {
	switch source {
	case "Services":
		return "Службы"
	case "Processes":
		return "Процессы"
	case "Startup":
		return "Автозагрузка"
	case "Scheduled tasks":
		return "Запланированные задачи"
	case "DNS cache":
		return "DNS-кэш"
	case "SRUDB/Data Usage":
		return "Использование данных"
	case "Current focused files":
		return "Текущие файлы"
	case "Minecraft logs/configs":
		return "Minecraft логи и конфиги"
	case "TEMP suspicious files":
		return "TEMP файлы"
	case "Browser artifacts":
		return "История браузера"
	case "Recycle Bin":
		return "Корзина"
	case "Recent":
		return "Недавние файлы"
	case "Prefetch":
		return "Prefetch (кэш запуска)"
	case "Manual files":
		return "Ручные файлы"
	case "Cheat data dirs":
		return "Директории читов"
	case "Installed programs":
		return "Установленные программы"
	default:
		return source
	}
}

func trim(s string, n int) string {
	s = strings.Join(strings.Fields(s), " ")
	if len(s) <= n {
		return s
	}
	return s[:n]
}

func hostname() string {
	h, _ := os.Hostname()
	return h
}

func mustWrite(path string, b []byte) {
	if err := os.WriteFile(path, b, 0644); err != nil {
		fmt.Println("Write failed:", err)
	}
}
