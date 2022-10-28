package render

type Options struct{}

var defaultOptions = Options{}

type Option func(*Options)
