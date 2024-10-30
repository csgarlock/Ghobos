
## Ghobos
Ghobos is a chess engine written in Go. The original goal of the project was to learn how to use go, but has since evolved into a much longer term project. It is hard to give it an accurate ELO rating at this point but my best guess right now would be 1800-2000. It is currently possible to play against Ghobos in the console.

#### Search Features
Ghobos currently has only very basic search features.
- The search is uses a negamax framework with alpha beta pruning
- Move ordering based on:
  - The best move from the transposition table
  - Internal iterative deepening if the transposition table does not have the current state
  - Attacks are ordered based on MVV-LVA
  - Quiet moves are ordered based on both the killer heuristic and the history heuristic
 - Late move reductions
 - Null move pruning

### Goals
#### Short Term Goals
1. Improve the static evaluation function
	- Add better understanding of pawns. Especially those that may promote
	- Add a metric for king safety
	- Fine tune parameters
2. Create a framework that allows to face different versions against each other to rigorously determine whether an update improves performance
3. Move generation, move making, and move unmaking code can be greatly improved in both speed and organization
4. Add more reductions, extensions, and pruning algorithms into the search, and improve the parameters of those that are already there

#### Long Term Goals
1. Implement a NNUE
