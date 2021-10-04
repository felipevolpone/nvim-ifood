# vim-ifood

[![demo](https://asciinema.org/a/439401.svg)](https://asciinema.org/a/439401?autoplay=1)

> Você ta programando e de repente bateu aquela fome? Sem problemas, peça seu Ifood sem sair do seu editor de texto favorito ❤️

Are you programming and suddenly your stomach is rumbling? No problem, order your Ifood without leaving your favorite text editor ❤️

## disclaimer

This is **obvisoully a joke**, even though it kinda works, I've made it just for fun :)

The code is terrible, because it's a **joke**, you got it?

The code is TERRIBLE, ok? really terrible. do not take any of this is consideration.

This is based on [Ifood](www.ifood.com.br) public APIs and it's not really possible to finish the order, but you can navigate around
and play with it.

## commands available

### `:IfoodLogin`
You start here. It'll ask your email and the OTP code sent to your inbox.

### `:IfoodAddress`
To pick an address again.

### `:IfoodHome`
To show the home cards.

### `:IfoodMerchants`
To list the merchants from the card chosen in the home.

## how?

Using [neovim](https://neovim.io) [Go client](https://github.com/neovim/go-client).

## building and running it
You can clone the repo, build the `go-ifood-lib` and clone the binary to the `vim-ifood` dir.
Then, create a ls or move the directory where the plugins are saved in your `nvim` config.

