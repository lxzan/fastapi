package fastapi

import (
	en "github.com/go-playground/locales/en_US"
	cn "github.com/go-playground/locales/zh_Hans_CN"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	en_translations "github.com/go-playground/validator/v10/translations/en"
	cn_translations "github.com/go-playground/validator/v10/translations/zh"
)

type Lang uint8

const (
	Chinese Lang = iota
	English
)

var (
	trans    ut.Translator
	uni      *ut.UniversalTranslator
	validate *validator.Validate
)

func init() {
	SetLang(Chinese)
}

func SetLang(lang Lang) {
	switch lang {
	case Chinese:
		setChinese()
	case English:
		setEnglish()
	}
}

func setChinese() {
	cn_translator := cn.New()
	uni = ut.New(cn_translator, cn_translator)
	trans, _ = uni.GetTranslator("zh")
	validate = validator.New()
	cn_translations.RegisterDefaultTranslations(validate, trans)
}

func setEnglish() {
	en_translator := en.New()
	uni = ut.New(en_translator, en_translator)
	trans, _ = uni.GetTranslator("en")
	validate = validator.New()
	en_translations.RegisterDefaultTranslations(validate, trans)
}
