# Terminology
Persistence: Each update creates a new version of the data structure.
All old versions can be queried and may also be updated.

Types of persistence
- Partial persistence
  - Updates operate only on latest version of structure
  - Queries can examine old versions
  - History is linear (sequence of operations forms a single timeline)
- Full persistence
  - Updates can be applied to any version
  - History forms a tree
(updating an old version creates a new branch)
- Confluent persistence
  - Updates can combine multiple versions (like git merge)
  - History forms a directed acyclic graph
# Requirements
## Scope
- Build a record system that can store different versions of a record and allow looking up the state of a record at a given timestamp
- Given the example of this problem, e.g. business owners change their physical address, partial persistence seems suitable as updates are likely to be serializable and a new update is likely to be applied on top of the most recent version
- Out of scope: support deleting a record or soft deleting a record version. Though these are important cases to handle, I think it's likely out of scope for this exercise.
## Requirements
### Functional
- F1 (out of scope): Support common data types: string, int, float,...
  - update: after reading the boilerplate, I can see that the data type is `string`. I'll mark this requirement as out of scope for now.
- F2: Users are able to create a record, insert more versions of the record
- F3: Users are able to retrieve a record at a given timestamp
- F4: Users are able to get a list of versions given a record
### Non-functional
- NF1: Persistent data storage with high availablity
- NF2: Consistency, patial persistence: the system is able to handle concurrent writes, keeping the linearability of versions of a record, i.e. a linear linked list without branches

# Acceptance criteria
1. F2: Able to create a record
- `POST /api/v2/record/{id}` will create a new record if the record ID is not present
2. F2: Able to modify an existing record (without losing its previous values)
- `POST /api/v2/record/{id}` will modify an existing record if the record ID is already present. Modifications include adding a new field, editing an existing field, and deleting a new field by setting that field's value to `null`
- subsequent `GET /api/v2/record/{id}` will return data that is the composition of the record's versions
2. F4: Users are able to get a list of versions given a record
- a new API endpoint `GET /api/v2/record/{record_id}/versions`
3. F3: Able to query for a particular record's values at different versions
- a new API endpoint `GET /api/v2/record/{record_id}/{timestamp}`, with timestamp being the parameter determining which versions decide the record's values at the time.
4. NF1: insert some records, turn off the computing units and turn them back on, should be able to retrieve previous records stored persistenly in SQLite.
5. (no longer needed, as we have chosen timestamp as version identifier, linearability is automatically guaranteed) NF2: after multiple insertions of different records and their versions, have a script to check the linearability of all records.

# Design considerations
A couple of questions:
1. How to define a version?

We can use a record's insertion timestamps as unique identifiers to distinguish its versions. To see a record's data, users need to provide a timestamp (could be a date picker in the UI), from which the record system finds the largest insertion timestamp that is less than this provided timestamp, and use it to come up with the final results.

2. How to effectively accumulate fields across a record's versions?
2.1 Merge updates one by one starting from the beginning at runtime

The most brute-force way is to store each update as a blob and merge appropriate updates at runtime. However, this approach could be quite time-consuming and memory-consuming if there are many updates.

records:
id | timestamp | data
---------------------
1  | ts1       | {"hello":"world"}
2  | ts2       | {"hello":"world 2","status":"ok"}
3  | ts3       | {"hello":null}


2.2 Calculate composition after each update

In addition to storing each update's data, we also calculate accumulated record data at that version.
Record retrieval at a particular version would be fast, with the trade-off of space. Obviously, if a record contains a lot of fields, a single field update would require storing another set of record.

records
id | timestamp | data | accumulated_data
---------------------
1  | ts1       | {"hello":"world"}                 | {"hello":"world"}
2  | ts2       | {"hello":"world 2","status":"ok"} | {"hello":"world 2","status":"ok"}
3  | ts3       | {"hello":null}                    | {"status":"ok"}



2.3 Hybrid: calculate composition after X updates, and apply the delta

We can combine the previous approaches. Depending on the traffic pattern, we could come up with a threshold, such as after 10 new updates, we calculate a checkpoint for the record, i.e. how the record looks like after 10, 20, 30... updates. If users request to see data at `version=13`, we find the checkpoint data at `version=10`, find the updates happening at timestamps between `10` and `13`, and merge their fields to return `{"hi":"mom"}`.

records
id  | timestamp | data | accumulated
---------------------
1   | ts1       | {"hello":"world"}    | NULL
...
10  | ts10       | {"hello":"world 2"} | {"hello":"world 2","status":"ok"}
11  | ts11       | {"hello":null}      | NULL
12  | ts11       | {"status":null}     | NULL
13  | ts11       | {"hi":"mom"}        | NULL

2.4 Flatten the data map, store field-value pair independently to quickly retrieve a field's latest value

The above approaches all require merging maps. Depending on the number of fields in a record and their uniqueness, if the number of fields is not too large, it may be more performant to store individual field-value pairs per update instead of keeping them in a blob.

records
id  | field_id | field_name |
-----------------------------
1   | 1        | hello
1   | 2        | status

customer_data
field_id | timestamp | field_value
----------------------------------
1        | ts1       | world
1        | ts2       | world 2
2        | ts2       | ok
1        | ts3       | NULL

To display a record's latest version, we need to know all of its fields over time, which is achievable from querying `records` table to get fields "hello" and "status".

Given a field, we can get its final value based on the latest insertion timestamp of that field, with a special case to handle, that is if the latest value of that field is NULL, we won't include the field in the final results. For example, the latest value of "hello" is NULL, and of "status" is "ok", meaning that the latest version should be `{"status":"ok"}`.

The same logics can be applied for displaying a record at any particular version. With a user-provided timestamp, for each field, we find its latest value being inserted just before the user-provided timestamp.

Basically, this SQL query would find the latest value of "hello" at any user-provided timestamp:
```
SELECT field_value
FROM customer_data
WHERE field_id = 1 AND timestamp < "2024-01-30"
ORDER BY timestamp DESC
LIMIT 1
```
The main cons of this approach is that the number of queries increases proportionally with the number of fields, which we may not have a control over - a bad actor could send a payload with lots of fields and hammer our database.

Conclusion: we can pick a strategy depending on the actual shape of the records and number of updates per record. Without these data, I'll blindly go with the approach 2.3 with an update interval of 10 versions.

# Implementation details
## Switch To Sqlite
- Add some code to create a database connection
- if there's no database, create one, and create `records` and `customer_data` table
- (if have time) handle database connection error

## Add Time Travel
- Add `/api/v2` endpoints
  - `POST /api/v2/record/{id}` create or update a record, return composition
  - `GET /api/v2/record/{id}` return the latest version
  - `GET /api/v2/record/{record_id}/versions` list all versions (insertion timestamps)
  - `GET /api/v2/record/{record_id}/{timestamp}` return composition at a particular version
- Testing 
  - acceptance criteria 1-4
Normally I write unit tests as I implement the code, but I'd need more time to get used to Go again, so I'll likely just use Postman or Python to visually check for expected results
  - acceptance criteria 5: kill the server and restart, see if a GET request still returns data
- Optimization for DB
  - modify primary keys, add indexes as needed