package helpers

func RemoveAccents(s string) string {
	accents := map[rune][]rune{
		'a': []rune("áàảãạăắằẳẵặâấầẩẫậ"),
		'e': []rune("éèẻẽẹêếềểễệ"),
		'i': []rune("íìỉĩị"),
		'o': []rune("óòỏõọôốồổỗộơớờởỡợ"),
		'u': []rune("úùủũụưứừửữự"),
		'y': []rune("ýỳỷỹỵ"),
		'd': []rune("đ"),
	}

	runesList := []rune(s)
	for i, r := range runesList {
		for unaccented, accentedChars := range accents {
			for _, ac := range accentedChars {
				if r == ac {
					runesList[i] = unaccented
					break
				}
			}
		}
	}
	return string(runesList)
}
