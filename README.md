# gopathdep

Manage your $GOPATH dependencies for your project.

To get a project setup with gopathdep, cd into the project's repo and record
its current dependencies: `gopathdep record`. This will create a
`pathdep.yaml` file at the root of the repo, which lists each dependency and
the desired commit. You can edit this file to specify a branch or tag, for
example: `branch: master` to always use the master branch of a particular
dependency, or `tag: v1.2` to always use the v1.2 tag.

You can check to see if your dependencies are what has been specified by doing
`gopathdep check`.

You can use `gopathdep apply` to apply your project's dependency requirements
to your $GOPATH. This will checkout packages to the specified
commit/tag/branch. If a dependency has modifications in it, gopathdep will
refuse to update that dependency and warn you about the inconsistency.

Note that this tool is not intended to work with the `vendor` directory. It is
intended to use your $GOPATH for this purpose instead. If you want to use the
vendor directory, I'd recommend using one of the other many dependency
management tools available.
