# TODO
- heartbeat
- client session (*)
    - session id
    - heartbeat timeout for session expiry

- watch event (*)
    - Since data changes must be approved by quorum/leader, data watches can be set globally and maintained locally at each server. 
    - when event fires, server clears the watch and notifies if the corresponding client is connected
    - watch event must be sent to client and response given before any other commands are processed for that client
    - Children changes are similar, but we maintain a separate list.
    - create(watch=true) => set child and data watch
    - getChildren(watch=true) => set child watch
    - getData(watch=true) and exists(watch=true) => set data watch
    - setData(watch=true) => trigger data watch
    - delete(path) => trigger data watch and child watch for path and child watch for its parent

- leader election protocol (*)
    - leader timeout
    - send out leader events

- parent counter for sequential nodes
    - monotonically increasing

- leader approved events
    - zxid
    - version numbers
    - ticks
    - realtime
    - zookeeper stat structure?

- znode data
    - sanity checks to ensure size is less than 1MB
    - size should be at most 1MB but usually much smaller

- ephemeral nodes (+)
    - delete on session expiry
- sequential nodes (*)

- client permissions (+)
    - read
    - write
    - create
    - delete
    - admin

- dynamic configuration (++)

