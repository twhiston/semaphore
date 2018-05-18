# Todays TODO and shiz

* how to manage deletes? do them last?
* write tests that actually write down some data to test database interaction
* dont pass in db or use it globally, make a channel

----------
The Why?

* semaphore is an app not a lib, structure it to reflect that (cobra in root)
* write as little code as possible
* better test coverage
* reduce error surface area to make it easier to manage

----------

* there is a need for a db layer, as we do db calls in the task runner. Currently commented out!
    this probably means that we need a further layer of abstraction to implement
* much clearer seperation of things. Probably means reworking a lot of the structure

Could you even seperate the api from the task runner?


DESIGN

api should be thin, prefering to delegate things out to other components lower down in the system, this should allow us to generate the api layer

What level is the db interaction on?

MIDDLEWARE
returned by function, build from generic wrapper and custom functions

GET
request -> router -> dbquery -> api result

router has the logic to make the request, and the get function is generic
Cannot be pure template as requires custom functions for selects and we need to write these in code


GOLDEN PATH

start refactoring the various parts and see what could be broken out and how