import { test } from 'node:test';
import assert from 'node:assert/strict';
import { startTarget } from '../src/start.js';

test('a new tab opens the local app, the server, or the choice', () => {
  assert.equal(startTarget({ base: '', mode: 'local' }), 'local');
  // Local mode wins even when an address was saved earlier.
  assert.equal(startTarget({ base: 'http://h:8090', mode: 'local' }), 'local');
  assert.equal(startTarget({ base: 'http://h:8090', mode: 'server' }), 'server');
  // Installs from before the local mode stored only an address.
  assert.equal(startTarget({ base: 'http://h:8090', mode: '' }), 'server');
  assert.equal(startTarget({ base: '', mode: '' }), 'choose');
  assert.equal(startTarget({ base: '', mode: 'server' }), 'choose');
});
