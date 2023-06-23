package store

/* Key: The key represents the unique identifier or name associated with the data being stored. It is used to retrieve or reference the data in the database.

Value: The value corresponds to the actual data that is being stored or persisted in the database.
It can be any binary or structured data that the application or user wants to store.

Metadata: Metadata provides additional information about the entry, such as timestamps,
version numbers, or flags. This information can be used for various purposes,
including data consistency checks, concurrency control, or tracking modifications.

Operation Type: The operation type indicates the nature of the operation being performed on the entry.
Common operation types include "create" (inserting a new key-value pair),
"update" (modifying an existing key-value pair), or "delete" (removing a key-value pair).

Transaction ID or Sequence Number: To ensure atomicity and consistency,
WAL entries often include a transaction ID or a sequence number.
This identifier helps in tracking the order of operations
and ensuring that changes are applied in the correct sequence during recovery or replication scenarios.

Additional Metadata or Flags: Depending on the specific database system and its requirements,
additional metadata or flags may be included in the WAL entry.
These can provide information for handling durability, replication, or other specific functionalities provided by the database.
*/

// If I am writing my own WAL Implementation , this is what will be required atleast
// lampost timestamp , I can write about it in the documentation as well - Use atomic operation for increment of this
// key
// value
// action
// metadata - What is the information that we should put in metadata.
// we can focus on this metadata when we start with the consensus algorithm

// for this we would need to add further file operations ?
// checkFileSize ?
// setToFile is already there .
// getAllDataOfFile
