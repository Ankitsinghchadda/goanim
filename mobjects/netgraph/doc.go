// Package netgraph provides labeled-circle nodes and connecting edges
// for graph-theory visualizations — social networks, microservice maps,
// state-machine-like topologies. Differs from systemdesign in that:
//
//   - Nodes are circles (graph-theory convention), not labeled boxes.
//   - Edges are configurable: directed (with arrowhead) or undirected.
//   - Edges support a "pulse" reveal — a brighter sub-segment travels
//     along the edge to show signal flow / cascading failure.
package netgraph
