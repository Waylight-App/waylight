// Package answers is the smart search on the server (ADR 0026): intents as
// data, tolerant matching of the words of a query, and answers as data in
// the language of the client. Modules contribute a Source each; the search
// module ranks the intents of all enabled modules and asks the sources of
// the best ones for an answer. The web interface and the mobile app render
// the same answers.
package answers
