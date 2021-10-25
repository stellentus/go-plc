// Package libplctag provides a connection to a physical PLC device using the libplctag library.
// The library is available at https://github.com/libplctag/libplctag.
// libplctag will link against is 2.1.0.
//
// If the package is built with the 'stub' build tag it will link against the system object 'libplctagstub' instead.
// The stub library is assumed to be an API compatible implementation with the real library, one is available at https://github.com/dijkstracula/plcstub.
package libplctag
