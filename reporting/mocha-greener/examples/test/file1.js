const assert = require('assert');

it('test pass number 1 ', () => {
    assert.equal(1, 1);
});

it('test fail number 1 ', () => {
    assert.equal(1, 2);
});

it('test err number 1 ', () => {
    throw new Error('test err number 1');
});

it.skip('test skip number 1 ', () => {
    throw new Error('test err number 1');
});

describe('describe number 1', () => {
    it('describe test 1', () => {
        assert.equal(1, 1);
    });

    describe('describe x2 number 1', () => {
        it('describe x2 test 1', () => {
            assert.equal(1, 1);
        });
    });
});