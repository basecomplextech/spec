

service Service {
    // Method doc comment.
    method(msg string) (msg1 string);

    // Method1 doc comment.
    method1(a int64, b string, c bool) (a1 int64, b1 string, c1 bool);

    // Service doc comment
    service(id bin128) (sub Subservice);
}

service Subservice {
    hello(msg string) (msg1 string);
}
