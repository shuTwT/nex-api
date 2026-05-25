export default {
  extends: ['@commitlint/config-conventional'],
  rules: {
    // 自定义限定提交类型，你可以根据团队需求增删
    'type-enum': [
      2,
      'always',
      [
        'feat',     // 新功能
        'fix',      // 修复 Bug
        'docs',     // 文档变更
        'style',    // 代码格式（不影响代码运行的变动）
        'refactor', // 代码重构
        'test',     // 增加测试
        'chore',    // 构建过程或辅助工具的变动
        'revert'    // 回退
      ]
    ],
    'subject-case': [0] // 不限制主题的大小写
  }
};