

service TestService {
    // Method doc comment.
    method(msg string 1) (msg string 1);

    // Method0 doc comment.
    method0();

    // Method1 doc comment.
    method1(
        a int       1, 
        b string    2, 
        c bool      3,
    ) (
        a int       1, 
        b string    2, 
        c bool      3,
    );

    // Method2 doc comment.
    method2(Request) Response;

    // Method3 doc comment.
    method3(Request) oneway;

    // Method10, 11, 12 have channels.
    method11(Request) (<-In) Response;
    method12(Request) (->Out) Response;
    method13(Request) (<-In ->Out) Response;

    // Subservice doc comment.
    subservice(id bin128 1) TestSubservice;
}

subservice TestSubservice {
    hello(msg string 1) (msg string 1);
}
