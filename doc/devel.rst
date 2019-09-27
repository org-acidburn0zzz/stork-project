.. _devel:

*****************
Developer's Guide
*****************

This part of the documentation pertains to some basic information intended to be used by developers.

Backend unit-tests
==================

We expect to have a very good test coverage for the code we produce, similar to what is there in
the Kea project. Each file whatever.go is supposed to be accompanied with whatever_test.go.
To run such unit-test, please use:

.. code-block:: console
   go test
