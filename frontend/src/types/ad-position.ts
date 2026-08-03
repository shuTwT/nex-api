export enum AdPosition {
  HOME_BANNER = "HOME_BANNER",
  SIDEBAR_TOP = "SIDEBAR_TOP",
  SIDEBAR_BOTTOM = "SIDEBAR_BOTTOM",
  CONTENT_TOP = "CONTENT_TOP",
  CONTENT_BOTTOM = "CONTENT_BOTTOM",
  POPUP = "POPUP",
}

export const AdPositionLabels: Record<AdPosition, string> = {
  [AdPosition.HOME_BANNER]: "首页横幅",
  [AdPosition.SIDEBAR_TOP]: "侧边栏顶部",
  [AdPosition.SIDEBAR_BOTTOM]: "侧边栏底部",
  [AdPosition.CONTENT_TOP]: "内容顶部",
  [AdPosition.CONTENT_BOTTOM]: "内容底部",
  [AdPosition.POPUP]: "弹窗广告",
};

export const AdPositionOptions = Object.entries(AdPositionLabels).map(
  ([value, label]) => ({
    value: value as AdPosition,
    label,
  }),
);
