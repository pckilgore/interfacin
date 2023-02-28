TODO: 

It also doesn't force the Serder convention on the model? Am I being too opinionated about that?
I do think it's quite nice for another model to be able to grab deserialization logic in
particular and cannonically re-use it, particularly in a relational database world where a service
might join rows of that model and wants a validated way of constructing it to do shit on it
without going request (and N + 1 crazy) on an ID basis.

This also brings up: How hidden should the capabilities of the peristance layer be within the 
model? Right now, we do a lot with the transaction object/database connection like tree shit and
footprints. Most of that isn't implemented in terms of a store and requires a database connection
and is tighly coupled to the specifics of that database. I think in that case, that tightly coupled
thing just becomes a dependency of service construction just like a store: at least then we're 
being honest about what we're doing when we implement a model.
