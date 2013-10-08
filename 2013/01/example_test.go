package main

func TestFirstParsePath(t *testing.T) {
	path := "D/go/code/../src/warcluster/tests/first/../../"

	if parsePath(path) != "/D/go/src/warcluster/" {
		t.Error("Result path is ", parsePath(path))
	}
}

func TestSecondParsePath(t *testing.T) {
	path := "python/movies/episode1/../../lectures/lecture1/examples/../code/../../../mostImportant/MonthyPython/quotes/../"

	if parsePath(path) != "/python/mostImportant/MonthyPython/" {
		t.Error("Result path is ", parsePath(path))
	}
