import * as React from "react";

const MOBILE_BREAKPOINT = 768;
const TABLET_BREAKPOINT = 1024;

export type DeviceType = "mobile" | "tablet" | "desktop";

interface DeviceInfo {
  deviceType: DeviceType;
  isMobile: boolean;
  isTablet: boolean;
  isDesktop: boolean;
  isTouchDevice: boolean; // mobile || tablet
}

/**
 * デバイスタイプを判定するフック
 * - mobile: < 768px
 * - tablet: 768px - 1023px
 * - desktop: >= 1024px
 */
export function useDevice(): DeviceInfo {
  const [deviceType, setDeviceType] = React.useState<DeviceType>("desktop");

  React.useEffect(() => {
    const updateDeviceType = () => {
      const width = window.innerWidth;
      if (width < MOBILE_BREAKPOINT) {
        setDeviceType("mobile");
      } else if (width < TABLET_BREAKPOINT) {
        setDeviceType("tablet");
      } else {
        setDeviceType("desktop");
      }
    };

    // 初期値を設定
    updateDeviceType();

    // リサイズイベントでデバイスタイプを更新
    window.addEventListener("resize", updateDeviceType);
    return () => window.removeEventListener("resize", updateDeviceType);
  }, []);

  const isMobile = deviceType === "mobile";
  const isTablet = deviceType === "tablet";
  const isDesktop = deviceType === "desktop";

  return {
    deviceType,
    isMobile,
    isTablet,
    isDesktop,
    isTouchDevice: isMobile || isTablet,
  };
}

/**
 * 後方互換のため維持（768px未満判定）
 */
export function useIsMobile() {
  const [isMobile, setIsMobile] = React.useState<boolean | undefined>(
    undefined,
  );

  React.useEffect(() => {
    const mql = window.matchMedia(`(max-width: ${MOBILE_BREAKPOINT - 1}px)`);
    const onChange = () => {
      setIsMobile(window.innerWidth < MOBILE_BREAKPOINT);
    };
    mql.addEventListener("change", onChange);
    setIsMobile(window.innerWidth < MOBILE_BREAKPOINT);
    return () => mql.removeEventListener("change", onChange);
  }, []);

  return !!isMobile;
}
