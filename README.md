# qmk-analyzer
## Description
A keyboard analyzer for custom layouts in the [QMK](https://config.qmk.fm/#/) format.
## Why?
A few months ago, I build a custom [ferris sweep](https://github.com/davidphilipbarr/Sweep) after after watching several of Ben Vallack's [videos on the keyboard](https://youtu.be/8wZ8FRwOzhU?si=WzESVyo3ULcMy9vJ). I put a slightly modified ISRT layout on it and began learning the new layout. After getting to the point where I could fairly consistently type letters, I started thinking about how to setup the rest of the symbols I would need, and wanted a way to analyze how efficient different keymaps would be. While keyboard analyzers do exist, most notably [this one](https://patorjk.com/keyboard-layout-analyzer/#/main), I wanted something a little more flexible. One of the main features of the QMK firmware is the ability to switch layers to reach keys, allowing you to keep more of your most common keys on your home row and strongest fingers, and none of the analyzers I could find supported that functionality. So, I had to build my own, and this project is the result.

Not all QMK features are supported yet (maybe ever), notably several methods for layer switching, because I don't use them and they add significantly to the complexity of how I designed the sequencer system. Because I wanted to support multiple layers, there is the potential for a single key to be present in multiple layers, or for a key to only be accessible from a specific layer. Therefore, the sequencer needs to not only figure out what finger will be used for the next key, but also if a layer switch needs to happen, or which layer to switch to if the next key is present in multiple layers, and whether the fingers used to get to that layer (if the layer is activated by holding a key), are the same that are needed to press the actual key, thereby excluding that as an option. 
## Quick Start
To build this project yourself, you must have at least v1.22 of [Go](https://go.dev/doc/install) installed.
Build and run the server file:
```bash
git clone https://github.com/jfosburgh/qmk-analyzer
cd qmk-analyzer
go build -o ./bin ./cmd/server/ 
./bin/server
```
Then navigate to [http://localhost:8080](http://localhost:8080)
## Usage
Once the server is running and you have navigated to the webpage, choose your keymap. This can be done by uploading a new keymap json as downloaded from [QMK Configurator](https://config.qmk.fm/#/), or selecting one from the dropdown menu. If a layout file for your keymap can't be found on the server's filesystem, you will be prompted to upload one. These files are expected to be in the info.json format for the layouts defined in the [QMK Firmware Repo](https://github.com/qmk/qmk_firmware/tree/master/layouts/default). Finally, you will be prompted to choose or create a fingermap, which tells the server what finger is used to press each key. If fingermaps already exist for your layout, those will also be available to select from. 

You are now set up to analyze your keyboard on your choice of text. Paste the text you would like analyzed in the text field, press analyze, and away you go. The text will be analyzed on your keymap, as well as any other keymaps that share the same layout, so you can compare with any other keymap that exists for your keyboard. The reported statistics are currently same finger bigrams (the same finger being used to press two keys in a row), total finger travel, and number of layer switches. These three components are then combined with equal weights to produce the overall score for your keyboard (the lower the better).

*Please note, rather than use cookies, on first load a session ID is created and stored in a hidden form, which is then sent down with every request made to keep track of your choices (thanks [HTMX](https://htmx.org/)). This means that sessions will not persist on a refresh!*
## Contributing
This tool in it's current state does everything I need it to do, so I have no current plans to continue development or evaluate/accept pull requests. If you have changes you'd like to make, I suggest forking the project and modifying it however you like.
