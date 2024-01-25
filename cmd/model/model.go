package model

const PluginSemVer = "v0.0.1"
const PluginName = "vision-plugin-10100-v1"
const PluginCommand = "10100"
const PluginOutputDir = "."

type PluginConfig struct {
	Modules []string `json:"module"`
}

type PluginData struct {
	PluginConfig PluginConfig `json:"10100"`
}
