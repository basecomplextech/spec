

service TestService {
    // Method doc comment.
    method(msg string 1) (msg string);

    // Method1 doc comment.
    method1(a int 1, b string 2, c bool 3) (a int, b string, c bool);

    // Subservice doc comment.
    subservice(id bin128 1) (sub TestSubservice);
}

subservice TestSubservice {
    hello(msg string 1) (msg string);
}
