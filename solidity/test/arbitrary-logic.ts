import chai from "chai";
import { ethers, network} from "hardhat";
import { solidity } from "ethereum-waffle";
import { TestLogicContract } from "../typechain/TestLogicContract";
import { IERC20} from "../typechain/IERC20"

import { deployContracts } from "../test-utils";
import {
  getSignerAddresses,
  makeCheckpoint,
  signHash,
  makeTxBatchHash,
  examplePowers
} from "../test-utils/pure";

chai.use(solidity);
const { expect } = chai;


async function runTest(opts: {
  // Issues with the tx batch
  batchNonceNotHigher?: boolean;
  malformedTxBatch?: boolean;

  // Issues with the current valset and signatures
  nonMatchingCurrentValset?: boolean;
  badValidatorSig?: boolean;
  zeroedValidatorSig?: boolean;
  notEnoughPower?: boolean;
  barelyEnoughPower?: boolean;
  malformedCurrentValset?: boolean;
}) {

    await network.provider.request({
      method: "hardhat_impersonateAccount",
      params: ["0x0c731fb0d03211dd32a456370ad2ec3ffad46520"]}
    )

  let lp_signer = await ethers.provider.getSigner("0x0c731fb0d03211dd32a456370ad2ec3ffad46520")


  // Prep and deploy contract
  // ========================
  const signers = await ethers.getSigners();
  const peggyId = ethers.utils.formatBytes32String("foo");
  // This is the power distribution on the Cosmos hub as of 7/14/2020
  let powers = examplePowers();
  let validators = signers.slice(0, powers.length);
  const powerThreshold = 6666;
  const {
    peggy,
    testERC20,
    checkpoint: deployCheckpoint
  } = await deployContracts(peggyId, validators, powers, powerThreshold);

  let usdc_eth_lp = (await ethers.getContractAt('IERC20', '0xb4e16d0168e52d35cacd2c6185b44281ec28c9dc', lp_signer)) as unknown as IERC20;


  const TestLogicContract = await ethers.getContractFactory("TestLogicContract");
  const logicContract = (await TestLogicContract.deploy(usdc_eth_lp.address)) as TestLogicContract;
  await logicContract.transferOwnership(peggy.address);

  console.log((await(usdc_eth_lp.balanceOf(lp_signer._address))).toString())
   
  await usdc_eth_lp.functions.approve(peggy.address, 10000)

  // await usdc_eth_lp.functions.transfer(peggy.address, 10000)

  // peggy.connect(lp_signer);
  let peggy_lp_signer = peggy.connect(lp_signer);


  // Transfer out to Cosmos, locking coins
  // =====================================
  // await testERC20.functions.approve(peggy.address, 10000);
  await peggy_lp_signer.functions.sendToCosmos(
    usdc_eth_lp.address,
    ethers.utils.formatBytes32String("myCosmosAddress"),
    1000
  );



  // Prepare batch
  // ===============================
  // This batch contains 10 transactions which each:
  // - Transfer 5 coins from Peggy's wallet to the logic contract
  // - Pay a fee of 1 coin
  // - Call transferTokens on the logic contract, transferring 2+2 coins to signer 20
  //
  // After the batch runs, signer 20 should have 40 coins, Peggy should have 940 coins,
  // and the logic contract should have 10 coins
  const numTxs = 10;
  const txLogicContractAddresses = new Array(numTxs);
  const txPayloads = new Array(numTxs);
  const txFees = new Array(numTxs);

  const txAmounts = new Array(numTxs);
  for (let i = 0; i < numTxs; i++) {
    txFees[i] = 1;
    txAmounts[i] = 10;
    txLogicContractAddresses[i] = logicContract.address;
    txPayloads[i] = logicContract.interface.encodeFunctionData("transferTokens", [await signers[20].getAddress(), 2, 2])
  }

  if (opts.malformedTxBatch) {
    // Make the fees array the wrong size
    txFees.pop();
  }

  let batchNonce = 1
  if (opts.batchNonceNotHigher) {
    batchNonce = 0
  }


  // Call method
  // ===========
  const methodName = ethers.utils.formatBytes32String(
    "logicBatch"
  );
  let abiEncoded = ethers.utils.defaultAbiCoder.encode(
    [
      "bytes32",
      "bytes32",
      "uint256[]",
      "address[]",
      "uint256[]",
      "bytes[]",
      "uint256",
      "address"
    ],
    [
      peggyId,
      methodName,
      txAmounts,
      txLogicContractAddresses,
      txFees,
      txPayloads,
      batchNonce,
      usdc_eth_lp.address
    ]
  );
  let digest = ethers.utils.keccak256(abiEncoded);
  let sigs = await signHash(validators, digest);
  let currentValsetNonce = 0;
  if (opts.nonMatchingCurrentValset) {
    // Wrong nonce
    currentValsetNonce = 420;
  }
  if (opts.malformedCurrentValset) {
    // Remove one of the powers to make the length not match
    powers.pop();
  }
  if (opts.badValidatorSig) {
    // Switch the first sig for the second sig to screw things up
    sigs.v[1] = sigs.v[0];
    sigs.r[1] = sigs.r[0];
    sigs.s[1] = sigs.s[0];
  }
  if (opts.zeroedValidatorSig) {
    // Switch the first sig for the second sig to screw things up
    sigs.v[1] = sigs.v[0];
    sigs.r[1] = sigs.r[0];
    sigs.s[1] = sigs.s[0];
    // Then zero it out to skip evaluation
    sigs.v[1] = 0;
  }
  if (opts.notEnoughPower) {
    // zero out enough signatures that we dip below the threshold
    sigs.v[1] = 0;
    sigs.v[2] = 0;
    sigs.v[3] = 0;
    sigs.v[5] = 0;
    sigs.v[6] = 0;
    sigs.v[7] = 0;
    sigs.v[9] = 0;
    sigs.v[11] = 0;
    sigs.v[13] = 0;
  }
  if (opts.barelyEnoughPower) {
    // Stay just above the threshold
    sigs.v[1] = 0;
    sigs.v[2] = 0;
    sigs.v[3] = 0;
    sigs.v[5] = 0;
    sigs.v[6] = 0;
    sigs.v[7] = 0;
    sigs.v[9] = 0;
    sigs.v[11] = 0;
  }

  await peggy.submitLogicBatch(
    await getSignerAddresses(validators),
    powers,
    currentValsetNonce,

    sigs.v,
    sigs.r,
    sigs.s,

    txAmounts,
    txLogicContractAddresses,
    txFees,
    txPayloads,
    batchNonce,
    testERC20.address
  );

  expect(
      (await testERC20.functions.balanceOf(await signers[20].getAddress()))[0].toNumber()
  ).to.equal(40);

  expect(
    (await testERC20.functions.balanceOf(peggy.address))[0].toNumber()
  ).to.equal(940);

  expect(
      (await testERC20.functions.balanceOf(logicContract.address))[0].toNumber()
  ).to.equal(10);
  
  expect(
    (await testERC20.functions.balanceOf(await signers[0].getAddress()))[0].toNumber()
  ).to.equal(9010);
}

describe.only("submitBatch tests", function () {
  it("throws on malformed current valset", async function () {
    await expect(runTest({ malformedCurrentValset: true })).to.be.revertedWith(
      "Malformed current validator set"
    );
  });

  it("throws on malformed txbatch", async function () {
    await expect(runTest({ malformedTxBatch: true })).to.be.revertedWith(
      "Malformed batch of transactions"
    );
  });

  it("throws on batch nonce not incremented", async function () {
    await expect(runTest({ batchNonceNotHigher: true })).to.be.revertedWith(
      "New batch nonce must be greater than the current nonce"
    );
  });

  it("throws on non matching checkpoint for current valset", async function () {
    await expect(
      runTest({ nonMatchingCurrentValset: true })
    ).to.be.revertedWith(
      "Supplied current validators and powers do not match checkpoint"
    );
  });


  it("throws on bad validator sig", async function () {
    await expect(runTest({ badValidatorSig: true })).to.be.revertedWith(
      "Validator signature does not match"
    );
  });

  it("allows zeroed sig", async function () {
    await runTest({ zeroedValidatorSig: true });
  });

  it("throws on not enough signatures", async function () {
    await expect(runTest({ notEnoughPower: true })).to.be.revertedWith(
      "Submitted validator set signatures do not have enough power"
    );
  });

  it("does not throw on barely enough signatures", async function () {
    await runTest({ barelyEnoughPower: true });
  });
});
