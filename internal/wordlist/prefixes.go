package wordlist

// BuiltinSuffixes are appended after the target name: target-SUFFIX
var BuiltinSuffixes = []string{
	"backup", "backups", "bak",
	"dev", "development", "staging", "stage",
	"prod", "production",
	"logs", "log",
	"data", "assets", "media",
	"static", "uploads", "files",
	"internal", "private", "public",
	"security", "sec",
	"test", "qa", "uat",
	"archive", "archives",
	"export", "exports", "import", "imports",
	"images", "img",
	"config", "configs", "conf",
	"store", "storage",
	"cache",
	"api", "api-v1", "api-v2",
	"web", "www", "cdn",
	"tmp", "temp",
	"docs", "documentation",
	"releases", "release",
	"build", "builds", "artifacts",
	"reports",
	"2023", "2024", "2025",
	"us-east-1", "us-west-2", "eu-west-1", "ap-southeast-1",
}

// BuiltinPrefixes are prepended before the target name: PREFIX-target
var BuiltinPrefixes = []string{
	"backup", "dev", "staging", "prod",
	"logs", "data", "assets", "media",
	"internal", "private", "public",
	"test", "qa",
	"archive", "cdn",
	"tmp", "temp",
	"old", "new",
}
