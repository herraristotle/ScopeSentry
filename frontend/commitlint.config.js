module.exports = {
  extends: ['@commitlint/config-conventional'],
  rules: {
    'type-enum': [
      2,
      'always',
      [
        'feat', // new feature
        'fix', // bug fix
        'docs', // documentation
        'style', // formatting/styles (changes that do not affect code behavior)
        'refactor', // refactoring (neither a new feature nor a bug fix)
        'perf', // performance improvements, e.g. speed or experience
        'test', // adding tests
        'ci', // continuous integration changes
        'chore', // build process or auxiliary tool changes
        'revert', // revert to a previous version
        'workflow', // workflow improvements
        'mod', // changes of uncertain category
        'wip', // work in progress
        'types', // type changes
        'release' // version release
      ]
    ],
    'subject-full-stop': [0, 'never'],
    'subject-case': [0, 'never']
  }
}
