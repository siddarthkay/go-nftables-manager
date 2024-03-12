# Go NFTables Manager

Go NFTables Manager is a Golang project that automates and manages `nftables` rules based on services registered in Consul. It retrieves the list of services from Consul and applies predefined firewall rules using nftables.

## Features

- Fetches services from Consul based on service name and filter criteria
- Applies firewall rules using `nftables` based on the retrieved services
- Creates and manages `nftables` sets for different environments (metrics, backups, app, logs)
- Supports retrying `Consul` API calls in case of failures
- Provides a test suite to validate the functionality of the `nftables` package

## Prerequisites

- `Go` programming language (version 1.20 or later)
- `nftables` installed on the system
- `Consul` server running and accessible

## Installation

1. Clone the repository:

   ```
   git clone https://github.com/siddarthkay/go-nftables-manager.git
   ```

2. Change to the project directory:

   ```
   cd go-nftables-manager
   ```

3. Build the project:

   ```
   go build
   ```

## Configuration

The project uses the following constants for configuration:

- `consulAddress`: The address of the Consul server (default: "http://localhost:8500")
- `serviceName`: The name of the service to fetch from Consul (default: "wireguard")
- `envValues`: An array of environment values (default: ["metrics", "logs", "backups", "app"])
- `stageValues`: An array of stage values (default: ["prod", "test"])

You can modify these constants in the `main.go` file to match your specific setup.

## Usage

1. Ensure that the Consul server is running and accessible.

2. Run the project:

   ```
   ./go-nftables-manager
   ```

   The project will fetch the services from Consul based on the configured service name and filter criteria, and apply the corresponding firewall rules using nftables.

3. Check the logs for any errors or success messages.

## Testing

The project includes a test suite for the `nftables` package. To run the tests, use the following command:

```
go test ./nftables
```

The tests use a sample `services.json` file located in the `testdata` directory to simulate the services retrieved from Consul.

## Contributing

Contributions are welcome! If you find any issues or have suggestions for improvement, please open an issue or submit a pull request.

## License

This project is licensed under the [MIT License](LICENSE).

## Acknowledgements

- The project uses the [nftables](https://netfilter.org/projects/nftables/) framework for managing firewall rules.
- It integrates with [Consul](https://www.consul.io/) for service discovery.

## Contact

For any questions or inquiries, please contact [siddarthkay@gmail.com](mailto:siddarthkay@gmail.com).