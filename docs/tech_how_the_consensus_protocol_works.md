# How the consensus protocol works

# Consensus Protocol

There are two protocols in `SEBAK`. One is `ISAAC` protocol for consensus and another is `Transaction Protocol` for sharing transactions.

## `ISAAC` Protocol
The `ISAAC` is a consensus protocol based on [`PBFT`](http://pmg.csail.mit.edu/papers/osdi99.pdf).
In the `ISAAC`, all nodes will vote two times to determine which transactions to include in this block.
For each voting unit, the nodes have the role `proposer` or` validator`.
The `proposer` gathers the transactions for this vote and puts them in a ballot and passes it to the `validator`.
The `validator` checks that the transactions in the ballot are valid and then votes.
If the voting determines that the ballot is valid, all nodes confirm the transactions in the ballot.

## State Diagram

                                     Height + 1
     +----------------------------------------------------------------------------+
     |                                                                            |
     v    `get ballot`         `get 67% of YES`           `get 67% of YES`        |
    INIT --------------> SIGN ------------------> ACCEPT ------------------> ALL_CONFIRM
      ^                   |                         |
      |   `get 34% of EXP`|         `get 34% of EXP`|
      +-------------------+-------------------------+
                       Round + 1

## `ISAAC` States
1. `INIT` - Proposed or received a ballot with transactions.

1. `SIGN` - Checked that the ballot is valid by itself.

1. `ACCEPT` - Checked that the ballot is valid by validators.

1. `ALL_CONFIRM` - Confirmed block with the transactions in the ballot.

## Voting Hole

1. `YES` - Agreed

1. `NO` - Disagreed

1. `EXP` - Expired

## Terms

* TIMEOUT_INIT - The timeout for `INIT` state. The default value is 2 sec.

* TIMEOUT_SIGN - The timeout for `SIGN` state. The default value is 2 sec.

* TIMEOUT_ACCEPT - The timeout for `ACCEPT` state. The default value is 2 sec.

* B(ISAAC State, Voting Hole) - A ballot with ISAAC state and voting hole.

### Examples

* B(`INIT`, `YES`) - A ballot with ISAAC state `INIT` and voting hole `YES`.

* B(`SIGN`, `NO`) - A ballot with ISAAC state `INIT` and voting hole `NO`.

* B(`ACCEPT`, `EXP`) - A ballot with ISAAC state `INIT` and voting hole `EXP`.

## Voting Process

### Network Start
* At the beginning of the network, the genesis block is saved with block height 1 and round 0.
* The node start with `INIT` state, height 2 and round 0.

### `INIT`
1. The timer is set to TIMEOUT_INIT.
1. The steps to propose transactions are as follows.
   * If the node is [proposer](./tech_how_to_select_the_proposer.md) of this round, it broadcasts a B(`INIT`, `YES`),
      * The ballot includes valid transactions(only hashes) or empty(If the transaction pool is empty).
      * When the node broadcasts the ballot, the node goes to the next state.
   * If the node is not `proposer`, it waits until receiving the B(`INIT`, `YES`).
      * When it receives the ballot, the node goes to the next state.
   * When the timer expires, the node goes to the next state.

### `SIGN`
1. The timer is reset to TIMEOUT_SIGN.
1. The node checks [the proposed ballot is valid](./tech_how_to_check_a_ballot_is_valid.md).
   * If the proposed ballot is valid(or empty), the node broadcasts B(`SIGN`, `YES`).
   * If the proposed ballot is invalid, the node broadcasts B(`SIGN`, `NO`).
   * When the timer expires, the node hasn't sent a ballot in the SIGN state yet, it broadcasts B(`SIGN`, `EXP`).
1. Each node receives ballots and goes to the next state after,
   * the number of B(`SIGN`, `YES`) is greater than or equal to 67% of validators, the node broadcasts B(`ACCEPT`, `YES`).
   * the number of B(`SIGN`, `NO`) or B(`SIGN`, `EXP`) is greater than 33% of validators,
      the node goes back to `INIT` state with same height and round + 1.

### `ACCEPT`
1. The timer is reset to TIMEOUT_ACCEPT.
1. When the timer expires, the node broadcasts B(`ACCEPT`, `EXP`).
1. Each node receives ballots and,
   * if the number of B(`ACCEPT`, `YES`) is greater than or equal to 67% of validators, the node goes to the next state.
   * if the number of B(`ACCEPT`, `NO`) or B(`ACCEPT`, `EXP`) is greater than 33% of validators, 
      the node goes back to `INIT` state with same height and round + 1.

### `ALL-CONFIRM`
1. the node confirms and saves the block with proposed transactions even though it is empty.
1. the node filters [invalid transactions](./tech_how_to_check_a_ballot_is_valid.md#transaction-validation) in the transaction pool.
1. It goes to `INIT` state with height + 1 and round 0.

### `INIT`
* If the node is `proposer`, it waits for [block time buffer](./tech_how_to_calculate_block_time_buffer.md) and propose.
* Else, it waits for `block time buffer` + TIMEOUT_INIT to receive proposed ballot.

## Transaction Protocol
1. [`Transaction Protocol`](./tech_how_transactions_are_shared.md) is a protocol for sharing transactions between the nodes.