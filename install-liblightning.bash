#!/bin/bash

function main {
    git clone https://github.com/lightning/liblightning.git
    cd liblightning
    make
    sudo make install
}

main "$@"

