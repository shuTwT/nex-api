import { Flex, Spin } from "antd";

interface ConsolePageLoadingProps {
  fullHeight?: boolean;
}

export function ConsolePageLoading({
  fullHeight = false,
}: ConsolePageLoadingProps) {
  return (
    <Flex
      align="center"
      justify="center"
      className={fullHeight ? "min-h-screen" : "min-h-80"}
    >
      <Spin size="large" delay={150} />
    </Flex>
  );
}
