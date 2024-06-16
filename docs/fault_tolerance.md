# Fault Tolerance

- [Fault Tolerance](#fault-tolerance)
- [Follower Faults](#follower-faults)
    - [Cases](#cases)
- [Leader Faults](#leader-faults)
- [Election Faults](#election-faults)

We tested our system with the following fault events

# Follower Faults
Expected behavior: 
- The leader should be able to continue processing requests without any issues.
- When the follower restarts, it should be up to date with the leader.

### Cases
1. While no requests are being processed
2. Follower faults before forwarding a request to the leader
   1. Client times out and retries
3. Follower faults after forwarding a request to the leader, before the request is committed
   1. Request will be delivered by all servers, but client will not receive a response


# Leader Faults
Expected behaviour:
- A new leader with the highest `LastZxid` should be elected
- When the old leader restarts, it should become a follower and be up to date with the new leader

# Election Faults
We assume that it is not possible for servers to fault while they are in the middle of an election.