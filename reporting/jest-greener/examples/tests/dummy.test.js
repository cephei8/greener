test.only('adds 1 + 2 to equal 3', () => {
    expect(1 + 2).toBe(3);
});

test.only('adds 1 + 2 to equal 4', () => {
    expect(1 + 2).toBe(4);
});

test.only('exception', () => {
    throw new Error('abc');
});

test('skipped test', () => {
    throw new Error('xyz');
});
