

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

    // Subservice doc comment.
    subservice(id bin128 1) (sub TestSubservice 1);
}

subservice TestSubservice {
    hello(msg string 1) (msg string 1);
}
