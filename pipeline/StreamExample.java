package pipeline;

import java.util.stream.Stream;

public class StreamExample {
	public static void main(String[] args) {
		Stream
				.of(1, 2, 3, 4, 5) // creating the source stream
				.filter(e -> e % 2 == 1) // T1: first operator
				.map(e -> e * 3) // T2: second operator
				.map(e -> e + 1) // T3: third operator
				.forEach(System.out::println); // Terminal operator

		// Output:
		// 4
		// 10
		// 16
	}
}
