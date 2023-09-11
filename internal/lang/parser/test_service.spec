

service TestService {
    // Method doc comment.
    method(msg string 1) (msg string 1);

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
    method2(Request) (Response);

    // Subservice doc comment.
    subservice(id bin128 1) (TestSubservice);
}

subservice TestSubservice {
    hello(msg string 1) (msg string 1);
}
