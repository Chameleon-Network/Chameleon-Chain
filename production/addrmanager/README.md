# AddrManager Package

## Overview

The `AddrManager` package is responsible for managing, storing, and handling known peer addresses in a decentralized network. This package provides functionality to store the addresses of connected peers persistently, allowing the application to retain this data between restarts. The core functionality of the `AddrManager` revolves around tracking peer addresses, saving them to disk periodically, and providing mechanisms to add, retrieve, and mark peers as "good."

### Key Features

1. **Persistent Address Storage**: The package supports saving known peer addresses to a file on disk. These addresses can be loaded and reused on application restart.
   
2. **Address Pool Management**: It maintains a dynamic pool of addresses, which can be marked as good when a successful connection is established.

3. **Periodic Saving**: The `AddrManager` periodically saves the address list to disk, ensuring that the state is preserved across application restarts.

4. **Graceful Shutdown**: When the application stops, the `AddrManager` ensures that any unsaved addresses are written to disk, preserving the most up-to-date state.

5. **Error Handling**: The package uses a robust error-handling system, allowing the system to respond gracefully to issues like file corruption or invalid address data.

6. **Concurrency Safe**: The `AddrManager` ensures thread-safety by using mutex locks when accessing and modifying the address pool, allowing it to operate in a concurrent environment without risking data corruption.

---

## Data Structures

### `AddrManager`

The `AddrManager` is the main struct that manages the addresses and is responsible for saving, loading, and updating peer information. It holds:
- A map (`addrIndex`) that stores known peer addresses indexed by their raw address.
- A `key` that uniquely identifies the address manager instance.
- A channel (`cQuit`) to signal when the manager should stop.
- Various synchronization primitives, like a mutex (`mtx`) for safe concurrent access.

### `serializedKnownAddress`

This struct represents the data for a single known address that is saved to disk. It includes:
- `Addr`: The raw address of the peer.
- `Src`: The peer’s source identifier, used to uniquely identify the peer.
- `PublicKey`: The public key of the peer.
- `PublicKeyType`: The type of the public key.

### `serializedAddrManager`

This struct represents the serialized version of the entire address manager, including metadata like:
- `Version`: The version of the serialized data.
- `Key`: The key associated with this instance of the address manager.
- `Addresses`: A list of `serializedKnownAddress` objects representing all known peers.

---

## Functionality

### Address Storage and Retrieval

The `AddrManager` supports two main operations for address management:
1. **Saving Known Addresses**: It periodically saves the list of known peer addresses to a file (e.g., `peer.json`). This process ensures that known peers are stored persistently, even if the application is restarted.
   
2. **Loading Known Addresses**: When the `AddrManager` starts, it attempts to load the list of known addresses from the file. If the file is corrupt or missing, the manager resets and starts fresh.

### Address Handling

- **Marking Addresses as Good**: When an address is successfully connected to (e.g., after a handshake or version exchange), it can be marked as "good" using the `Good` method. This allows the manager to track which peers are reachable and stable.

- **Address Pool**: The package maintains a pool of addresses, which can be accessed using the `AddressCache` method. The addresses in this pool are stored in a map, where the raw address is the key and the `peer.Peer` object is the value. This map allows quick access and efficient updates.

### Graceful Shutdown

Upon shutdown, the `AddrManager` ensures that any changes to the address pool (such as new or updated addresses) are saved to disk before the application exits. This is done through the `Stop` method, which gracefully shuts down the manager and saves all known addresses.

---

## Error Handling

The `AddrManager` package uses a custom error handling system to track and report errors. Each error is associated with a unique error code and message, making it easier to debug issues that arise in the address management process. Some of the errors handled include:
- **File Errors**: Issues related to opening, encoding, or decoding the peer data file.
- **Version Mismatch**: If the stored address data has an unsupported version, the manager will throw an error.
- **Unexpected Errors**: Any unforeseen errors are captured with a generic error code.

---

## Conclusion

The `AddrManager` package is a crucial component of this decentralized network system, handling the storage and management of peer addresses in a persistent and efficient manner. It ensures that known peers are available even after a restart and allows for periodic saving of the address pool. With support for graceful shutdowns, concurrency, and error handling, the `AddrManager` offers a reliable and flexible solution for managing network peers.

