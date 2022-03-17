package spec

type ListWriter[W any] struct {
	w     *Writer
	begin func(*Writer) W
}

func BeginList[W any](w *Writer, begin func(*Writer) W) ListWriter[W] {
	w.BeginList()

	return ListWriter[W]{
		w:     w,
		begin: begin,
	}
}

func (w ListWriter[W]) BeginNext() W {
	return w.begin(w.w)
}

func (w ListWriter[W]) EndNext() error {
	return w.w.Element()
}

// Value

type ValueListWriter[T any] struct {
	w     *Writer
	write func(el T) error
}

func BeginValueList[T any](w *Writer, write func(T) error) ValueListWriter[T] {
	w.BeginList()

	return ValueListWriter[T]{
		w:     w,
		write: write,
	}
}

func (w ValueListWriter[T]) Next(el T) error {
	w.write(el)
	return w.w.Element()
}
