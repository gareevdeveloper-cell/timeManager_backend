package slug

import (
	"regexp"
	"strconv"
	"strings"
	"unicode"
)

// cyrillicToLatin — базовая транслитерация кириллицы в латиницу.
var cyrillicToLatin = map[rune]string{
	'а': "a", 'б': "b", 'в': "v", 'г': "g", 'д': "d", 'е': "e", 'ё': "e",
	'ж': "zh", 'з': "z", 'и': "i", 'й': "y", 'к': "k", 'л': "l", 'м': "m",
	'н': "n", 'о': "o", 'п': "p", 'р': "r", 'с': "s", 'т': "t", 'у': "u",
	'ф': "f", 'х': "h", 'ц': "ts", 'ч': "ch", 'ш': "sh", 'щ': "sch",
	'ъ': "", 'ы': "y", 'ь': "", 'э': "e", 'ю': "yu", 'я': "ya",
	'А': "a", 'Б': "b", 'В': "v", 'Г': "g", 'Д': "d", 'Е': "e", 'Ё': "e",
	'Ж': "zh", 'З': "z", 'И': "i", 'Й': "y", 'К': "k", 'Л': "l", 'М': "m",
	'Н': "n", 'О': "o", 'П': "p", 'Р': "r", 'С': "s", 'Т': "t", 'У': "u",
	'Ф': "f", 'Х': "h", 'Ц': "ts", 'Ч': "ch", 'Ш': "sh", 'Щ': "sch",
	'Ъ': "", 'Ы': "y", 'Ь': "", 'Э': "e", 'Ю': "yu", 'Я': "ya",
}

var slugCleanup = regexp.MustCompile(`[^a-z0-9]+`)

// From создаёт URL-дружественный slug из строки.
// Нормализация: транслитерация кириллицы, lowercase, пробелы → дефисы.
func From(s string) string {
	var b strings.Builder
	for _, r := range s {
		if replacement, ok := cyrillicToLatin[r]; ok {
			b.WriteString(replacement)
		} else if unicode.IsLetter(r) || unicode.IsNumber(r) {
			b.WriteRune(unicode.ToLower(r))
		} else if r == ' ' || r == '-' || r == '_' {
			b.WriteByte('-')
		}
	}
	result := slugCleanup.ReplaceAllString(strings.Trim(b.String(), "-"), "-")
	result = strings.Trim(result, "-")
	if result == "" {
		return "org"
	}
	return result
}

// WithSuffix добавляет суффикс к slug для разрешения коллизий (slug-2, slug-3, ...).
func WithSuffix(base string, n int) string {
	if n <= 1 {
		return base
	}
	return base + "-" + strconv.Itoa(n)
}
