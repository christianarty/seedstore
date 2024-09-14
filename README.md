# Seedstore

Seedstore is a tool inspired by [Queue4Download](https://github.com/weaselBuddha/Queue4Download) that helps manage and automate your seedbox downloads to your local devices. However, the difference with this is that we have binaries to make it as easy as possible.

## Table of Contents

- [Seedstore](#seedstore)
  - [Table of Contents](#table-of-contents)
  - [Overview](#overview)
  - [Features](#features)
  - [Setup](#setup)
  - [Usage](#usage)
    - [Example Commands](#example-commands)
  - [Configuration](#configuration)
  - [Contributing](#contributing)
  - [License](#license)

## Overview

Seedstore is designed to simplify the management of your seedbox by automating tasks such as downloading, uploading, and organizing files. It leverages MQTT for event-driven operations and provides a flexible rule-based system for handling different scenarios.

## Features

- **MQTT Integration**: Seamlessly connect to your MQTT server for event-driven operations.
- **Rule-Based Processing**: Define custom rules to handle different scenarios and automate tasks.
- **LFTP Support**: Efficiently transfer files using LFTP with configurable threads and segments.
- **Flexible Configuration**: Easily configure the tool using a JSON file.

## Setup

To get started with Seedstore, you need to have a `config.json` located in `$HOME/.seedstore`. Below is the schema for the configuration file:

```json5
{
  mqtt: {
    username: "freerealestate", // MQTT user account name
    password: "foobar", // MQTT user account password
    host: "192.168.3.2", // MQTT server for events
  },
  server: {
    defaultCode: "V", // If the processing of rules fails, this is the default code that is assigned
    codeConditions: [
      {
        value: "loremipsum",
        operator: "in",
        entity: "name",
        code: "A",
      },
    ],
  },
  client: {
    codeDestinations: {
      A: "/local/path/on/client/for/code",
      V: "/local/path/on/client/for/default/code",
      // make sure your have the default code specified here
    },
    lftp: {
      threads: 5, // the amount of threads to use on LFTP transfer
      segments: 4, // the amount of segments to use when mirroring directories on LFTP transfer
    },
    serverInfo: {
      host: "192.168.1.2", //  the ip address to connect to seedbox
      username: "iliketurtles", // the username for the seedbox
      password: "batquot", // the password for the seedbox
    },
  },
}
```

## Usage

To use Seedstore, follow these steps:

1. **Install Dependencies**: Ensure you have Go installed and run `go mod tidy` to install dependencies.
2. **Build the Project**: Run `go build -v ./...` to build the project.
3. **Run the Application**: Execute the binary with the appropriate commands.

### Example Commands

- **Publish**: Publish a message to the MQTT server, it will take your params and send a JSON-formatted message to the specified topic (default:"queue")
  ```bash
  ./seedstore publish --name "example" --hash "12345" --location "/path/to/file" --category "movies" --topic "queue"
  ```
- **Subscribe**: On the client device, you can subscript to a topic on the MQTT server.

```bash
./seedstore subscribe --topic "queue"
```

## Configuration

The configuration file is located at `$HOME/.seedstore/config.json`. Make sure to update the file with your specific settings as specified above.

## Contributing

We welcome contributions! Please follow these steps to contribute:

1. Fork the repository.
2. Create a new branch (`git checkout -b feature-branch`).
3. Make your changes.
4. Commit your changes (`git commit -am 'Add new feature'`).
5. Push to the branch (`git push origin feature-branch`).
6. Create a new Pull Request.

## License

This project is licensed under the MIT License. See the [LICENSE](./LICENSE) file for details.
