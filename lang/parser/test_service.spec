

service TestService {
    // Method doc comment.
    method(msg string) (msg string);

    // Method1 doc comment.
    method1(a int, b string, c bool) (a int, b string, c bool);

    // Service doc comment
    service(id bin128) (sub TestSubservice);
}

service TestSubservice {
    hello(msg string) (msg string);
}
