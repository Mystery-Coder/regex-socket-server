package main

var regexQuestions = []string{
	"\\d{3}\\w",
	"^[a-c]{2}\\d{2}$",
	"^(cat|dog)\\d$",
	"^[A-Z]{3}-\\d{2}$",
	"^[a-z]{4}\\d$",
	"^\\d{2}:[a-z]{3}$",
	"^(red|blue|green)$",
	"^\\w{5}$",
	"^\\d{4}-\\d{2}-\\d{2}$",
	"^(hi|yo){2}$",
	"^[a-f0-9]{3}$",
	"^file_\\d{1,3}$",
	"^\\+?\\d{1,2}-\\d{3}$",
	"^[a-z]{2}\\.[a-z]{2}$",
	"^(up|down)-[LR]$",
	"^[^aeiou]{3}\\d$",
}

var stringQuestions = [][]string{
	{"aa", "bb", "cc"},
	{"ab1", "bc2", "cd3"},
	{"cat-01", "dog-12", "fox-99"},
	{"ABC_123", "XYZ_007", "QWE_900"},
	{"a-bb", "c-dd", "e-ff"},
	{"12:ab", "34:cd", "56:ef"},
	{"red", "blue", "green"},
	{"v1.2.3", "v2.0.0", "v10.4.7"},
	{"2024-01-05", "1999-12-31", "2030-07-19"},
	{"aaab", "bbbc", "cccd"},
	{"xYz9", "aBc0", "mNo7"},
	{"file_1a", "file_2b", "file_9z"},
	{"0x1f", "0x2a", "0x9c"},
	{"ha-ha", "ha-ha-ha", "ha-ha-ha-ha"},
	{"pre_cat", "post_dog", "pre_fox"},
}
