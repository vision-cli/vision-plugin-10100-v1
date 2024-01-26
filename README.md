# ![logo](./images/vision-logo.svg "Vision") &nbsp; Vision Plugin - 10.10.0

<p align="center">
  <img width="150" src="./images/10100.svg" />
</p>

This plugin creates a starter 10.10.0 app. Read about 10100 apps in the [book](https://atos-digital.github.io/10100-book) and see the [live demo](https://10100.digital)

Vision plugins require golang (https://go.dev) to be installed

## Install

Install the plugin with

```
go install github.com/vision-cli/vision-plugin-10100-v1@latest
```

You will now see the plugin help in the vision cli

```
vision 10100 --help
```

You will now see the plugin available in vision's help

```
vision --help
```

The plugin will be listed in plugins list

```
vision plugins list
```

## Run

Create a 10100 project

```
vision init <project name>
```

Switch to the project directory

```
cd <project name>
```

Initialise the 10100 plugin

```
vision 10100 init <module name>
```
