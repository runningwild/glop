The ai package provides a way of having a program run user-code at runtime.  This is done using two separate packages, yedparse and polish.

    yedparse: github.com/runningwild/yedparse
    polish: github.com/runningwild/Polish


An ai script is supplied as a yed graph saved as a .xgml file.  The format is as follows:

 * There must be exactly one start node, specified by being labeled with the text 'start', this node should have exactly one edge leading to another node.  It is this node at which the graph will begin evaluating.
 * Each node should have a single expression in polish notation that will be evaluated using the polish package.  The functions available to the expression should be known a-priori.
 * Each node may have any number of output edges colored black.
 * Each node may have output edges colored either red or green, but if there are any red edges there must be at least one green edge and vice-versa.
 * If there are any red or green output edges from a node then when that node's expression is evaluated it must evaluate to either true or false.  If it evaluates to true then an edge is randomly selected among all output edges that are colored either green or black.  If it evaluates to false then an edge is randomly selected among all output edges that are colored either red or black.  The selected edge is followed to find the next node to evaluate.
 * If there are no output edges that are colored green or red then when that node's expression is evaluated it does not need to evaluate to anything.  After evaluating an edge is randomly selected out of all output edges (which are implicitly colored black), and that edge is followed to find the next node to evaluate.


![sample ai graph](/runningwild/go-glop/raw/master/ai/sample.png)

The following functions and variables are used in the above graph:

    numVisibleEnemies
    distBetween
    nearestEnemy
    attack
    advanceTowards
    done
    me

There are two nodes with red and green output edges, their expressions both evaluate to a boolean value.  There are two nodes (other than the start node) with only black output edges, these nodes do not evaluate to boolean values, but that is ok since the result of the expression will never be checked since there are no red or green output edges.
