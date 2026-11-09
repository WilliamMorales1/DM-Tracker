import { test, expect } from 'vitest';
import { isCloudModel, modelKey, normalizeModels, pickModel } from './models';

// Real /api/models output from each server (trimmed).
const lemonade = ['Gemma-4-E4B', 'Qwen3.5-4B', 'ollama.gemma4:31b', 'ollama.nemotron-3-super'];
const ollama = ['qwen3.5:4b', 'gemma4:31b-cloud', 'nemotron-3-super:cloud'];

test('isCloudModel recognizes Ollama and Lemonade cloud ids', () => {
  expect(isCloudModel('gemma4:31b-cloud')).toBe(true);
  expect(isCloudModel('nemotron-3-super:cloud')).toBe(true);
  expect(isCloudModel('ollama.gemma4:31b')).toBe(true);
  expect(isCloudModel('Qwen3.5-4B')).toBe(false);
  expect(isCloudModel('qwen3.5:4b')).toBe(false);
});

test('modelKey matches same model across servers', () => {
  expect(modelKey('ollama.gemma4:31b')).toBe(modelKey('gemma4:31b-cloud'));
  expect(modelKey('ollama.nemotron-3-super')).toBe(modelKey('nemotron-3-super:cloud'));
  expect(modelKey('qwen3.5:latest')).toBe(modelKey('qwen3.5'));
  expect(modelKey('qwen3.5:4b')).not.toBe(modelKey('qwen3.5:8b'));
});

test('pickModel defaults to cloud model on either server', () => {
  expect(pickModel(lemonade, null)).toBe('ollama.gemma4:31b');
  expect(pickModel(ollama, null)).toBe('gemma4:31b-cloud');
});

test('pickModel falls back to first model when no cloud model', () => {
  expect(pickModel(['Qwen3.5-4B', 'Gemma-4-E4B'], null)).toBe('Qwen3.5-4B');
  expect(pickModel([], null)).toBe('');
  expect(pickModel([], 'gemma4:31b-cloud')).toBe('');
});

test('pickModel keeps saved model when present', () => {
  expect(pickModel(lemonade, 'Gemma-4-E4B')).toBe('Gemma-4-E4B');
  expect(pickModel(ollama, 'qwen3.5:4b')).toBe('qwen3.5:4b');
});

test('pickModel carries saved model across server switch', () => {
  expect(pickModel(lemonade, 'nemotron-3-super:cloud')).toBe('ollama.nemotron-3-super');
  expect(pickModel(ollama, 'ollama.nemotron-3-super')).toBe('nemotron-3-super:cloud');
});

test('pickModel ignores saved model missing from current server', () => {
  expect(pickModel(lemonade, 'qwen3.5:4b')).toBe('ollama.gemma4:31b');
});

test('normalizeModels tolerates null and junk responses', () => {
  expect(normalizeModels(null)).toEqual([]);
  expect(normalizeModels({ data: [] })).toEqual([]);
  expect(normalizeModels(['a', 1, '', null, 'b'])).toEqual(['a', 'b']);
});
