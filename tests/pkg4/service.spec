

service Service {
    // Method doc comment.
    method(msg string) (msg string);

    // Method1 doc comment.
    method1(a int64, b string, c bool) (a int64, b string, c bool);

    // Service doc comment
    service(id bin128) (sub Subservice);
}

service Subservice {
    hello(msg string) (msg string);
}
