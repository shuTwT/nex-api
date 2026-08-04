import { useEffect, useState } from "react";
import {
  Alert,
  Avatar,
  Button,
  Card,
  Col,
  Descriptions,
  Input,
  Modal,
  Row,
  Space,
  Statistic,
  Tag,
  Typography,
} from "antd";
import { api, responseData } from "@/lib/api";
import {
  CreditCard,
  Activity,
  Coins,
  Plus,
  Wallet,
  Ticket,
  Gift,
  AlertTriangle,
} from "lucide-react";
import { RechargeDialog } from "@/components/recharge-dialog";
import { ConsolePageLoading } from "@/components/console-page-loading";
import { toast } from "sonner";

interface UserProfile {
  id: string;
  name: string | null;
  email: string;
  image: string | null;
  username: string;
  role: string;
  credits: number;
  createdAt: string;
  totalCreditsSpent: number;
  totalRequests: number;
}

export default function PersonalPage() {
  const [profile, setProfile] = useState<UserProfile | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [rechargeDialogOpen, setRechargeDialogOpen] = useState(false);
  const [redeemInput, setRedeemInput] = useState("");
  const [isRedeeming, setIsRedeeming] = useState(false);
  const [confirmDialog, setConfirmDialog] = useState<{
    open: boolean;
    type: string;
    planName: string | null;
    credits: number | null;
  }>({ open: false, type: "", planName: null, credits: null });
  async function loadProfile() {
    setIsLoading(true);
    const result = await api.personal_profile_route_get();
    const data = responseData<UserProfile>(result);
    if (result.success && data) setProfile(data);
    setIsLoading(false);
  }
  useEffect(() => {
    loadProfile();
  }, []);
  async function handleRedeem() {
    if (!redeemInput.trim()) return;
    setIsRedeeming(true);
    const result = await api.personal_redeem_lookup_route_post({
      code: redeemInput,
    });
    const data = responseData<{
      type: string;
      planName: string | null;
      credits: number | null;
    }>(result);
    if (result.success && data)
      setConfirmDialog({
        open: true,
        type: data.type,
        planName: data.planName,
        credits: data.credits,
      });
    else toast.error(result.error || "查询失败");
    setIsRedeeming(false);
  }
  async function handleConfirmRedeem() {
    setIsRedeeming(true);
    const result = await api.personal_redeem_route_post({ code: redeemInput });
    if (result.success) {
      toast.success("兑换成功");
      setRedeemInput("");
      setConfirmDialog({
        open: false,
        type: "",
        planName: null,
        credits: null,
      });
      loadProfile();
    } else toast.error(result.error || "兑换失败");
    setIsRedeeming(false);
  }
  if (isLoading) return <ConsolePageLoading />;
  const displayName = profile?.name || profile?.username;
  const statsCards = [
    {
      title: "历史消耗",
      value: profile?.totalCreditsSpent || 0,
      icon: <Activity size={20} />,
    },
    {
      title: "请求次数",
      value: profile?.totalRequests || 0,
      icon: <CreditCard size={20} />,
    },
  ];
  return (
    <div className="flex flex-col gap-6">
      <div>
        <Typography.Title level={2}>个人中心</Typography.Title>
        <Typography.Text type="secondary">
          查看和管理您的个人信息
        </Typography.Text>
      </div>
      <RechargeDialog
        open={rechargeDialogOpen}
        onOpenChange={setRechargeDialogOpen}
      />
      <Row gutter={[24, 24]}>
        <Col span={24}>
          <Card>
            <Space size="large" wrap>
              <Avatar size={96} src={profile?.image}>
                {displayName?.charAt(0).toUpperCase() || "U"}
              </Avatar>
              <div>
                <Typography.Title level={3}>{displayName}</Typography.Title>
                <Space wrap>
                  <Tag>{profile?.email}</Tag>
                  <Tag color={profile?.role === "admin" ? "purple" : "default"}>
                    {profile?.role === "admin" ? "管理员" : "普通用户"}
                  </Tag>
                  <Tag>
                    ID: {profile?.id}
                  </Tag>
                  <Tag>注册时间:{profile?.createdAt
                    ? new Date(profile.createdAt).toLocaleDateString("zh-CN")
                    : "-"}</Tag>
                </Space>
              </div>
            </Space>
          </Card>
        </Col>
      </Row>
      <Row gutter={[24, 24]}>
        <Col xs={24} md={9}>
          <Card
            title={
              <span className="flex items-center gap-2">
                <Wallet size={20} />
                钱包
              </span>
            }
          >
            <Statistic
              title="当前余额"
              value={profile?.credits || 0}
              suffix="积分"
              prefix={<Coins size={20} />}
            />
            <Button
              type="primary"
              icon={<Plus size={16} />}
              className="mt-4"
              onClick={() => setRechargeDialogOpen(true)}
            >
              充值
            </Button>
            <Space.Compact className="mt-4 w-full">
              <Input
                placeholder="请输入兑换码"
                value={redeemInput}
                onChange={(event) => setRedeemInput(event.target.value)}
                onPressEnter={handleRedeem}
                disabled={isRedeeming}
              />
              <Button
                type="primary"
                icon={<Ticket size={16} />}
                loading={isRedeeming}
                disabled={!redeemInput.trim()}
                onClick={handleRedeem}
              >
                兑换
              </Button>
            </Space.Compact>
          </Card>
        </Col>
        <Col xs={24} md={15}>
          <Card title="统计信息" className="h-full">
            <Row gutter={[16, 16]}>
              {statsCards.map((stat) => (
                <Col span={12} key={stat.title}>
                  <Card size="small">
                    <Statistic
                      title={stat.title}
                      value={stat.value}
                      prefix={stat.icon}
                    />
                  </Card>
                </Col>
              ))}
            </Row>
          </Card>
        </Col>
      </Row>
      <Modal
        open={confirmDialog.open}
        title={
          <span className="flex items-center gap-2">
            <AlertTriangle size={20} />
            确认兑换
          </span>
        }
        okText="确认兑换"
        cancelText="取消"
        confirmLoading={isRedeeming}
        onOk={handleConfirmRedeem}
        onCancel={() =>
          setConfirmDialog((previous) => ({ ...previous, open: false }))
        }
      >
        <Typography.Paragraph>
          您即将使用兑换码，请确认以下信息：
        </Typography.Paragraph>
        {confirmDialog.type === "subscription" ? (
          <Descriptions
            items={[
              {
                key: "type",
                label: "类型",
                children: (
                  <Tag color="blue" icon={<Gift size={12} />}>
                    订阅兑换码
                  </Tag>
                ),
              },
              {
                key: "plan",
                label: "订阅计划",
                children: confirmDialog.planName || "-",
              },
            ]}
          />
        ) : (
          <Descriptions
            items={[
              {
                key: "type",
                label: "类型",
                children: (
                  <Tag color="purple" icon={<Coins size={12} />}>
                    额度兑换码
                  </Tag>
                ),
              },
              {
                key: "credits",
                label: "获得额度",
                children: `${(confirmDialog.credits || 0).toLocaleString()} 积分`,
              },
            ]}
          />
        )}
        <Alert
          className="mt-4"
          type="warning"
          showIcon
          message="兑换后不可撤销，确认要继续吗？"
        />
      </Modal>
    </div>
  );
}
