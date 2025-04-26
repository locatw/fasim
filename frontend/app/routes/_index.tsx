import { Link } from "@remix-run/react";
import type { MetaFunction } from "@remix-run/node";

export const meta: MetaFunction = () => {
  return [
    { title: "Factory Automation Simulator (Fasim) - ホーム" },
    { name: "description", content: "生産自動化を行うゲームの生産プロセスをシミュレートしてグラフィカルに表示するアプリケーション" },
  ];
};

export default function Index() {
  return (
    <div className="container mx-auto p-4">
      <h1 className="text-3xl font-bold mb-6">Factory Automation Simulator (Fasim)</h1>

      <div className="mb-8">
        <p className="mb-4">
          生産自動化を行うゲームの生産プロセスをシミュレートしてグラフィカルに表示するアプリケーションです。
        </p>
        <p className="mb-4">
          Factorio、Satisfactory、Dyson Sphere Programなどの生産ラインをシミュレートし、
          リソースの流れや生産効率をグラフィカルに表示するためのツールです。
        </p>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
        <div className="border p-4 rounded shadow">
          <h2 className="text-xl font-semibold mb-2">生産レシピの定義</h2>
          <p>ゲーム内の生産レシピを定義し、管理します。</p>
          <Link to="/recipes" className="text-blue-600 hover:underline mt-2 inline-block">
            レシピを管理する →
          </Link>
        </div>

        <div className="border p-4 rounded shadow">
          <h2 className="text-xl font-semibold mb-2">生産ラインのシミュレーション</h2>
          <p>定義した生産ラインの動作をシミュレートします。</p>
          <Link to="/simulation" className="text-blue-600 hover:underline mt-2 inline-block">
            シミュレーションを開始する →
          </Link>
        </div>

        <div className="border p-4 rounded shadow">
          <h2 className="text-xl font-semibold mb-2">リソースフローの可視化</h2>
          <p>リソースの流れをグラフィカルに表示します。</p>
          <Link to="/visualization" className="text-blue-600 hover:underline mt-2 inline-block">
            可視化を表示する →
          </Link>
        </div>
      </div>
    </div>
  );
}
