"use client";

import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { 
  Receipt, 
  Download, 
  Filter,
  Calendar,
  DollarSign,
  CreditCard,
  FileText,
  ArrowUpRight
} from "lucide-react";

const billingStats = [
  {
    title: "本月消费",
    value: "¥ 128.50",
    change: "+12.5%",
    icon: DollarSign,
  },
  {
    title: "账户余额",
    value: "¥ 256.80",
    change: "-3.1%",
    icon: CreditCard,
  },
  {
    title: "待付款账单",
    value: "2",
    change: "",
    icon: FileText,
  },
  {
    title: "累计消费",
    value: "¥ 1,234.50",
    change: "",
    icon: Receipt,
  },
];

const bills = [
  {
    id: "INV-2024-001",
    date: "2024-03-01",
    period: "2024年2月",
    amount: "¥ 89.50",
    status: "paid",
    items: 12,
  },
  {
    id: "INV-2024-002",
    date: "2024-02-01",
    period: "2024年1月",
    amount: "¥ 128.30",
    status: "paid",
    items: 15,
  },
  {
    id: "INV-2024-003",
    date: "2024-01-01",
    period: "2023年12月",
    amount: "¥ 67.20",
    status: "paid",
    items: 8,
  },
  {
    id: "INV-2024-004",
    date: "2024-03-15",
    period: "2024年3月",
    amount: "¥ 156.80",
    status: "pending",
    items: 18,
  },
];

const transactions = [
  {
    id: 1,
    type: "recharge",
    description: "账户充值",
    amount: "+¥ 200.00",
    time: "2024-03-10 14:23",
    balance: "¥ 256.80",
  },
  {
    id: 2,
    type: "consume",
    description: "API 调用消费",
    amount: "-¥ 12.50",
    time: "2024-03-10 12:15",
    balance: "¥ 56.80",
  },
  {
    id: 3,
    type: "consume",
    description: "API 调用消费",
    amount: "-¥ 8.30",
    time: "2024-03-09 18:45",
    balance: "¥ 69.30",
  },
  {
    id: 4,
    type: "refund",
    description: "退款",
    amount: "+¥ 15.00",
    time: "2024-03-08 10:20",
    balance: "¥ 77.60",
  },
  {
    id: 5,
    type: "consume",
    description: "API 调用消费",
    amount: "-¥ 23.40",
    time: "2024-03-07 16:30",
    balance: "¥ 62.60",
  },
];

export default function BillingPage() {
  return (
    <div className="space-y-6">
      {/* Page Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-slate-900">账单</h1>
          <p className="text-slate-500 mt-1">管理您的账单和支付记录</p>
        </div>
        <div className="flex gap-2">
          <Button variant="outline" size="sm" className="gap-2 cursor-pointer">
            <Calendar className="h-4 w-4" />
            选择日期
          </Button>
          <Button className="gap-2 cursor-pointer">
            <CreditCard className="h-4 w-4" />
            充值
          </Button>
        </div>
      </div>

      {/* Stats Grid */}
      <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-4">
        {billingStats.map((stat) => {
          const Icon = stat.icon;
          return (
            <Card key={stat.title} className="hover:shadow-md transition-shadow cursor-pointer">
              <CardContent className="p-6">
                <div className="flex items-center justify-between">
                  <div className="h-12 w-12 rounded-lg bg-blue-50 flex items-center justify-center">
                    <Icon className="h-6 w-6 text-blue-600" />
                  </div>
                  {stat.change && (
                    <div className={`flex items-center gap-1 text-sm ${
                      stat.change.startsWith("+") ? "text-green-600" : "text-red-600"
                    }`}>
                      <ArrowUpRight className="h-4 w-4" />
                      <span>{stat.change}</span>
                    </div>
                  )}
                </div>
                <div className="mt-4">
                  <p className="text-sm text-slate-500">{stat.title}</p>
                  <p className="text-2xl font-bold text-slate-900 mt-1">{stat.value}</p>
                </div>
              </CardContent>
            </Card>
          );
        })}
      </div>

      {/* Bills and Transactions */}
      <div className="grid gap-6 lg:grid-cols-2">
        {/* Bills */}
        <Card>
          <CardHeader>
            <div className="flex items-center justify-between">
              <CardTitle className="text-lg">账单列表</CardTitle>
              <Button variant="outline" size="sm" className="gap-2 cursor-pointer">
                <Download className="h-4 w-4" />
                导出
              </Button>
            </div>
          </CardHeader>
          <CardContent>
            <div className="space-y-4">
              {bills.map((bill) => (
                <div 
                  key={bill.id} 
                  className="flex items-center justify-between py-3 border-b border-slate-100 last:border-0 hover:bg-slate-50 transition-colors cursor-pointer rounded-lg px-2 -mx-2"
                >
                  <div className="flex-1 min-w-0">
                    <div className="flex items-center gap-2">
                      <p className="text-sm font-medium text-slate-900">{bill.id}</p>
                      <Badge 
                        variant="outline" 
                        className={
                          bill.status === "paid" 
                            ? "bg-green-50 text-green-700 border-green-200" 
                            : "bg-orange-50 text-orange-700 border-orange-200"
                        }
                      >
                        {bill.status === "paid" ? "已支付" : "待支付"}
                      </Badge>
                    </div>
                    <p className="text-xs text-slate-500 mt-1">{bill.period}</p>
                  </div>
                  <div className="text-right">
                    <p className="text-sm font-bold text-slate-900">{bill.amount}</p>
                    <p className="text-xs text-slate-500 mt-1">{bill.date}</p>
                  </div>
                </div>
              ))}
            </div>
          </CardContent>
        </Card>

        {/* Transactions */}
        <Card>
          <CardHeader>
            <div className="flex items-center justify-between">
              <CardTitle className="text-lg">交易记录</CardTitle>
              <Button variant="outline" size="sm" className="gap-2 cursor-pointer">
                <Filter className="h-4 w-4" />
                筛选
              </Button>
            </div>
          </CardHeader>
          <CardContent>
            <div className="space-y-4">
              {transactions.map((transaction) => (
                <div 
                  key={transaction.id} 
                  className="flex items-center justify-between py-3 border-b border-slate-100 last:border-0"
                >
                  <div className="flex-1 min-w-0">
                    <p className="text-sm font-medium text-slate-900">{transaction.description}</p>
                    <p className="text-xs text-slate-500 mt-1">{transaction.time}</p>
                  </div>
                  <div className="text-right">
                    <p className={`text-sm font-bold ${
                      transaction.type === "consume" ? "text-red-600" : "text-green-600"
                    }`}>
                      {transaction.amount}
                    </p>
                    <p className="text-xs text-slate-500 mt-1">余额: {transaction.balance}</p>
                  </div>
                </div>
              ))}
            </div>
          </CardContent>
        </Card>
      </div>
    </div>
  );
}
