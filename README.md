# Go NetCat Implementation

A Go implementation of the NetCat (`nc`) utility, focusing on networking functionality with a clean, idiomatic Go approach.

## Overview

This project aims to recreate the core functionality of the popular NetCat utility in Go. The TCP client component is the first implementation, with more features planned.

## Current Features

- TCP client implementation with:
  - Customizable input/output streams
  - Clean, minimalistic API
  - Standard library only (no external dependencies)

## Installation

Since this project is not yet published as a Go package, you can install it directly from GitHub:

```bash
# Clone the repository
git clone git@github.com:gppmad/gonc.git
cd gonc

# Build the project
go build -o nc
```

## Usage Examples

# Connect to a remote server
```bash
./gonc example.com 80
```


