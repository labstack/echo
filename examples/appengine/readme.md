# AppEngine Example

This is the website example adapted to demonstrate Echo being used with Google
AppEngine.

AppEngine provides serving of static files, request logging and http
compression so these are offloaded from the app. AppEngine also has it's own
web serving mechanism which requires Echo to be plugged into the root handler
of the default http router rather than being Run in main().

To run a local AppEngine development instance (assuming the SDK is installed):

  goapp serve

If tbe application id was changed to one you own you can deploy it with:

  goapp deploy

Conditional compilation still allows running the app standalone or on another
platform though where the Echo middlewares and static handlers are enabled.

Running the standalone version simply involves the normal build / execution:

  go build -o example
  ./example

Both versions are set to run at http://localhost:8080 (the default for the
AppEngine development server).
